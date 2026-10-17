package payment

import (
	"context"
	"database/sql"
	"fmt"
	billingsvc "pay-langgan/internal/services/billing"
	"time"

	provider "pay-langgan/pkg/payment"
)

type RecurringProvider interface {
	CreateRecurringCharge(context.Context, provider.RecurringChargeRequest) (*provider.RecurringChargeResponse, error)
	GetRecurringCharge(context.Context, string) (*provider.RecurringChargeResponse, error)
}

// Off by default. This controls NEW debits only; existing debits are reconciled
// even when the switch is turned off or customer consent is later revoked.
func (s *CheckoutService) EnableAutopay(enabled bool) { s.autopayEnabled = enabled }

type chargeWork struct {
	ID         int64
	BusinessID string
	Provider   string
	Request    provider.RecurringChargeRequest
}

func (s *CheckoutService) claimAutopay(ctx context.Context) (*chargeWork, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var row struct {
		BusinessID string    `db:"business_id"`
		Provider   string    `db:"provider"`
		InvoiceID  int       `db:"invoice_id"`
		Amount     float64   `db:"amount"`
		MethodID   int       `db:"payment_method_id"`
		Token      string    `db:"token"`
		MaxAmount  int64     `db:"max_amount"`
		ExpiresAt  time.Time `db:"consent_expires_at"`
		Cycle      string    `db:"billing_cycle"`
	}
	err = tx.GetContext(ctx, &row, `SELECT c.business_id,bg.provider,i.id AS invoice_id,i.amount,a.payment_method_id,pm.token,a.max_amount,a.consent_expires_at,s.billing_cycle
 FROM subscriptions s JOIN subscription_autopay a ON a.subscription_id=s.id
 JOIN customers c ON c.id=s.customer_id JOIN business_payment_gateways bg ON bg.business_id=c.business_id
 JOIN payment_methods pm ON pm.id=a.payment_method_id AND pm.customer_id=c.id AND pm.provider=bg.provider
 JOIN invoices i ON i.subscription_id=s.id
 WHERE (s.status='active' OR (s.status='pending' AND s.trial_ends_at IS NOT NULL AND s.trial_ends_at<=(NOW() AT TIME ZONE 'UTC')))
 AND a.enabled=TRUE AND a.consent_expires_at>(NOW() AT TIME ZONE 'UTC')
 AND s.billing_cycle IN ('monthly','yearly') AND i.status='pending' AND i.due_date<=(NOW() AT TIME ZONE 'UTC')
 AND i.period_start>=a.effective_at AND i.period_end>i.period_start
 AND NOT EXISTS(SELECT 1 FROM autopay_charges ac WHERE ac.invoice_id=i.id)
 AND NOT EXISTS(SELECT 1 FROM checkout_sessions cs WHERE cs.invoice_id=i.id AND cs.status IN ('creating','active'))
 ORDER BY i.due_date,i.id LIMIT 1 FOR UPDATE OF s,a,i SKIP LOCKED`)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	ref, err := secureToken(16)
	if err != nil {
		return nil, err
	}
	ref = "autopay_" + ref
	var paymentID int
	if err := tx.QueryRowContext(ctx, `INSERT INTO payments(invoice_id,provider,amount,status,created_at) VALUES($1,$2,$3,'pending',(NOW() AT TIME ZONE 'UTC')) RETURNING id`, row.InvoiceID, row.Provider, row.Amount).Scan(&paymentID); err != nil {
		return nil, err
	}
	var chargeID int64
	if err := tx.QueryRowContext(ctx, `INSERT INTO autopay_charges(invoice_id,payment_id,payment_method_id,reference_id,status)
 VALUES($1,$2,$3,$4,'creating') RETURNING id`, row.InvoiceID, paymentID, row.MethodID, ref).Scan(&chargeID); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO payment_attempts(payment_id,invoice_id,provider,attempt_number,status,attempted_at,created_at)
 VALUES($1,$2,$3,1,'pending',(NOW() AT TIME ZONE 'UTC'),(NOW() AT TIME ZONE 'UTC'))`, paymentID, row.InvoiceID, row.Provider); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE checkout_sessions SET status='expired',updated_at=(NOW() AT TIME ZONE 'UTC') WHERE invoice_id=$1 AND status IN ('created','failed')`, row.InvoiceID); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	amount := int64(row.Amount)
	work := &chargeWork{ID: chargeID, BusinessID: row.BusinessID, Provider: row.Provider, Request: provider.RecurringChargeRequest{ReferenceID: ref, PaymentTokenID: row.Token, Amount: amount,
		ReturnURL: s.returnBaseURL, ConsentExpiresAt: row.ExpiresAt, FrequencyDays: 28}}
	if row.Cycle == "yearly" {
		work.Request.FrequencyDays = 365
	}
	if float64(amount) != row.Amount || amount < 5000 || amount > 200000000 || amount > row.MaxAmount {
		if err := s.failAutopay(ctx, chargeID, "amount_outside_consent", true); err != nil {
			return nil, err
		}
		work.Request.PaymentTokenID = "" // Do not send this candidate to the provider.
	}
	return work, nil
}

// This transaction is separate from the provider call: a process crash cannot
// erase the fact that a charge may already have been sent.
func (s *CheckoutService) failAutopay(ctx context.Context, id int64, code string, definitive bool) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var paymentID int
	status := "review_required"
	if definitive {
		status = "failed"
	}
	err = tx.QueryRowContext(ctx, `UPDATE autopay_charges SET status=$2,failure_code=$3,updated_at=(NOW() AT TIME ZONE 'UTC'),
 next_check_at=(NOW() AT TIME ZONE 'UTC')+INTERVAL '5 minutes' WHERE id=$1 AND status NOT IN ('succeeded','failed') RETURNING payment_id`, id, status, code).Scan(&paymentID)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	if definitive {
		if _, err := tx.ExecContext(ctx, `UPDATE payments SET status='failed' WHERE id=$1 AND status='pending'`, paymentID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE payment_attempts SET status='failed',error_message=$2 WHERE payment_id=$1 AND status='pending'`, paymentID, code); err != nil {
			return err
		}
		var invoiceID int
		if err := tx.GetContext(ctx, &invoiceID, `SELECT invoice_id FROM payments WHERE id=$1`, paymentID); err != nil {
			return err
		}
		var pending bool
		if err := tx.GetContext(ctx, &pending, `SELECT status='pending' FROM invoices WHERE id=$1`, invoiceID); err != nil {
			return err
		}
		if pending {
			if err := billingsvc.QueuePaymentFailure(ctx, tx, invoiceID, code); err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}

func (s *CheckoutService) applyAutopay(ctx context.Context, id int64, response *provider.RecurringChargeResponse) error {
	if response == nil || response.PaymentRequestID == "" {
		return fmt.Errorf("missing payment request identity")
	}
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var invoiceID int
	if err := tx.GetContext(ctx, &invoiceID, `SELECT invoice_id FROM autopay_charges WHERE id=$1`, id); err != nil {
		return err
	}
	var invoiceStatus string
	if err := tx.GetContext(ctx, &invoiceStatus, `SELECT status FROM invoices WHERE id=$1 FOR UPDATE`, invoiceID); err != nil {
		return err
	}
	var row struct {
		PaymentID int     `db:"payment_id"`
		Ref       string  `db:"reference_id"`
		RequestID *string `db:"provider_request_id"`
		Amount    float64 `db:"amount"`
		Status    string  `db:"status"`
	}
	if err := tx.GetContext(ctx, &row, `SELECT ac.payment_id,ac.reference_id,ac.provider_request_id,p.amount,ac.status FROM autopay_charges ac
 JOIN payments p ON p.id=ac.payment_id WHERE ac.id=$1 FOR UPDATE OF ac`, id); err != nil {
		return err
	}
	if response.ReferenceID != row.Ref || response.Currency != "IDR" || float64(response.Amount) != row.Amount ||
		(row.RequestID != nil && *row.RequestID != response.PaymentRequestID) {
		return fmt.Errorf("payment request does not match charge")
	}
	if row.Status == "succeeded" {
		return nil
	}
	status, code, paymentStatus := "pending", "", "pending"
	switch response.Status {
	case "SUCCEEDED":
		if response.PaymentID == "" || (invoiceStatus != "pending" && invoiceStatus != "paid") {
			return fmt.Errorf("payment requires review")
		}
		status, paymentStatus = "succeeded", "success"
		if _, err := tx.ExecContext(ctx, `UPDATE invoices SET status='paid',paid_at=COALESCE(paid_at,(NOW() AT TIME ZONE 'UTC')) WHERE id=$1`, invoiceID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE payments SET status='success',provider_transaction_id=$2,paid_at=COALESCE(paid_at,(NOW() AT TIME ZONE 'UTC')) WHERE id=$1`, row.PaymentID, response.PaymentID); err != nil {
			return err
		}
		if err := billingsvc.ActivatePaidSubscription(ctx, tx, invoiceID); err != nil {
			return err
		}
		if err := billingsvc.QueuePaymentConfirmation(ctx, tx, invoiceID); err != nil {
			return err
		}
	case "FAILED", "EXPIRED", "CANCELED":
		status, paymentStatus = "failed", "failed"
		code = safeFailureCode(response.FailureCode)
		if _, err := tx.ExecContext(ctx, `UPDATE payments SET status='failed' WHERE id=$1 AND status='pending'`, row.PaymentID); err != nil {
			return err
		}
	case "REQUIRES_ACTION":
		status, code = "review_required", "customer_action_required"
	case "ACCEPTING_PAYMENTS", "AUTHORIZED":
	default:
		status, code = "review_required", "provider_status_requires_review"
	}
	if _, err := tx.ExecContext(ctx, `UPDATE autopay_charges SET provider_request_id=$2,status=$3,failure_code=NULLIF($4,''),
 updated_at=(NOW() AT TIME ZONE 'UTC'),next_check_at=(NOW() AT TIME ZONE 'UTC')+INTERVAL '2 minutes' WHERE id=$1`, id, response.PaymentRequestID, status, code); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE payment_attempts SET status=$2,error_message=NULLIF($3,'') WHERE payment_id=$1`, row.PaymentID, paymentStatus, code); err != nil {
		return err
	}
	if status == "failed" && invoiceStatus == "pending" {
		if err := billingsvc.QueuePaymentFailure(ctx, tx, invoiceID, code); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func safeFailureCode(code string) string {
	switch code {
	case "INSUFFICIENT_BALANCE", "INVALID_TOKEN", "EXPIRED_CARD", "CARD_DECLINED", "DECLINED_BY_ISSUER", "ISSUER_UNAVAILABLE", "USER_DID_NOT_AUTHORIZE", "SUSPECTED_FRAUDULENT", "STOLEN_CARD":
		return code
	default:
		return "payment_failed"
	}
}

func (s *CheckoutService) runAutopay(ctx context.Context, limit int, result *MaintenanceResult) error {
	if !s.autopayEnabled && s.businessGateways == nil {
		if _, ok := s.provider.(RecurringProvider); !ok {
			return nil
		}
	}
	for i := 0; i < limit; i++ {
		var row struct {
			ID         int64   `db:"id"`
			BusinessID string  `db:"business_id"`
			Provider   string  `db:"provider"`
			RequestID  *string `db:"provider_request_id"`
		}
		err := s.db.GetContext(ctx, &row, `WITH candidate AS (SELECT id FROM autopay_charges WHERE status IN ('creating','pending','review_required')
 AND next_check_at<=(NOW() AT TIME ZONE 'UTC') AND updated_at<(NOW() AT TIME ZONE 'UTC')-INTERVAL '30 seconds'
				 ORDER BY next_check_at,id LIMIT 1 FOR UPDATE SKIP LOCKED)
	 UPDATE autopay_charges ac SET next_check_at=(NOW() AT TIME ZONE 'UTC')+INTERVAL '2 minutes'
		 FROM candidate WHERE ac.id=candidate.id RETURNING ac.id,ac.provider_request_id`)
		if err == sql.ErrNoRows {
			break
		}
		if err != nil {
			return err
		}
		if err := s.db.QueryRowxContext(ctx, `SELECT c.business_id,p.provider FROM autopay_charges ac JOIN payments p ON p.id=ac.payment_id
			JOIN invoices i ON i.id=ac.invoice_id JOIN subscriptions sub ON sub.id=i.subscription_id JOIN customers c ON c.id=sub.customer_id WHERE ac.id=$1`, row.ID).Scan(&row.BusinessID, &row.Provider); err != nil {
			return err
		}
		result.AutopayChecked++
		if row.RequestID == nil {
			result.ReviewRequired++
			if err := s.failAutopay(ctx, row.ID, "charge_outcome_uncertain", false); err != nil {
				return err
			}
			continue
		}
		p, err := s.recurringProviderForBusiness(ctx, row.BusinessID, row.Provider)
		if err != nil {
			result.Failures++
			result.ReviewRequired++
			if err := s.failAutopay(ctx, row.ID, "gateway_configuration_unavailable", false); err != nil {
				return err
			}
			continue
		}
		response, err := p.GetRecurringCharge(ctx, *row.RequestID)
		if err == nil {
			err = s.applyAutopay(ctx, row.ID, response)
		}
		if err != nil {
			result.Failures++
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if err := s.failAutopay(ctx, row.ID, "reconciliation_failed", false); err != nil {
				return err
			}
		}
	}
	if !s.autopayEnabled {
		return nil
	}
	for i := 0; i < limit; i++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		work, err := s.claimAutopay(ctx)
		if err != nil {
			return err
		}
		if work == nil {
			break
		}
		if work.Request.PaymentTokenID == "" {
			result.Failures++
			continue
		}
		p, err := s.recurringProviderForBusiness(ctx, work.BusinessID, work.Provider)
		if err != nil {
			result.Failures++
			if saveErr := s.failAutopay(ctx, work.ID, "gateway_configuration_unavailable", true); saveErr != nil {
				return saveErr
			}
			continue
		}
		if work.Provider == "midtrans" {
			// Midtrans status lookup uses order_id. Persist it before calling
			// Charge API so an uncertain network result can be reconciled safely.
			if _, err := s.db.ExecContext(ctx, `UPDATE autopay_charges SET provider_request_id=$2 WHERE id=$1 AND provider_request_id IS NULL`, work.ID, work.Request.ReferenceID); err != nil {
				return err
			}
		}
		result.AutopayStarted++
		response, chargeErr := p.CreateRecurringCharge(ctx, work.Request)
		saveCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		if chargeErr != nil {
			definitive := provider.SessionCreationRejected(chargeErr)
			code := "charge_outcome_uncertain"
			if definitive {
				code = "charge_rejected"
			}
			err = s.failAutopay(saveCtx, work.ID, code, definitive)
			result.Failures++
		} else {
			// Keep the verified request ID even if settlement fails locally.
			if response != nil && response.ReferenceID == work.Request.ReferenceID && response.PaymentRequestID != "" {
				_, err = s.db.ExecContext(saveCtx, `UPDATE autopay_charges SET provider_request_id=$2 WHERE id=$1 AND (provider_request_id IS NULL OR provider_request_id=$2)`, work.ID, response.PaymentRequestID)
			}
			if err == nil {
				err = s.applyAutopay(saveCtx, work.ID, response)
			}
			if err != nil {
				result.Failures++
				err = s.failAutopay(saveCtx, work.ID, "settlement_requires_review", false)
			}
		}
		cancel()
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *CheckoutService) recurringProviderForBusiness(ctx context.Context, businessID, expectedProvider string) (RecurringProvider, error) {
	if s.businessGateways == nil {
		if expectedProvider != "" && expectedProvider != "xendit" {
			return nil, fmt.Errorf("legacy payment gateway does not support %s recurring payments", expectedProvider)
		}
		p, ok := s.provider.(RecurringProvider)
		if !ok {
			return nil, fmt.Errorf("provider does not support recurring payments")
		}
		return p, nil
	}
	gateway, err := s.gatewayForBusiness(ctx, businessID)
	if err != nil {
		return nil, err
	}
	if expectedProvider != "" && gateway.Config.Provider != expectedProvider {
		return nil, fmt.Errorf("configured payment gateway changed from %s", expectedProvider)
	}
	switch gateway.Config.Provider {
	case "xendit":
		p, createErr := provider.NewXenditProvider(gateway.Config.XenditSecretKey, gateway.Config.XenditCallbackToken, "https://api.xendit.co", false)
		if createErr != nil {
			return nil, createErr
		}
		return p, nil
	case "midtrans":
		origin := "https://api.sandbox.midtrans.com"
		if gateway.Config.Environment == "production" {
			origin = "https://api.midtrans.com"
		}
		return provider.NewMidtransProvider(gateway.Config.MidtransServerKey, gateway.Config.MidtransClientKey, origin, false), nil
	default:
		return nil, fmt.Errorf("business gateway does not support recurring payments")
	}
}

// Operator recovery for a charge whose response was lost. Never resend a debit.
func (s *CheckoutService) RecoverAutopay(ctx context.Context, id int64, requestID string) error {
	if requestID == "" {
		return fmt.Errorf("provider payment request ID required")
	}
	var businessID string
	if err := s.db.GetContext(ctx, &businessID, `SELECT c.business_id FROM autopay_charges ac JOIN invoices i ON i.id=ac.invoice_id
 JOIN subscriptions sub ON sub.id=i.subscription_id JOIN customers c ON c.id=sub.customer_id WHERE ac.id=$1`, id); err != nil {
		return err
	}
	var expectedProvider string
	if err := s.db.GetContext(ctx, &expectedProvider, `SELECT p.provider FROM autopay_charges ac JOIN payments p ON p.id=ac.payment_id WHERE ac.id=$1`, id); err != nil {
		return err
	}
	p, err := s.recurringProviderForBusiness(ctx, businessID, expectedProvider)
	if err != nil {
		return err
	}
	response, err := p.GetRecurringCharge(ctx, requestID)
	if err != nil {
		return err
	}
	return s.applyAutopay(ctx, id, response)
}
