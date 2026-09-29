package platformbilling

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	billing "pay-langgan/internal/services/billing"
	provider "pay-langgan/pkg/payment"
)

type BatchResult struct {
	InvoicesCreated int `json:"invoices_created"`
	SessionsChecked int `json:"sessions_checked"`
	AutopayStarted  int `json:"autopay_started"`
	AutopayChecked  int `json:"autopay_checked"`
	ReviewRequired  int `json:"review_required"`
	Failures        int `json:"failures"`
}

// RunOnce uses one claim per transaction. Provider calls are outside database locks.
func (s *Service) RunOnce(ctx context.Context, limit int) (BatchResult, error) {
	var result BatchResult
	if limit < 1 || limit > 1000 {
		return result, fmt.Errorf("limit must be 1..1000")
	}
	for i := 0; i < limit; i++ {
		created, err := s.renewOne(ctx)
		if err != nil {
			return result, err
		}
		if !created {
			break
		}
		result.InvoicesCreated++
	}
	var sessions []struct {
		ID        int64  `db:"id"`
		SessionID string `db:"provider_session_id"`
	}
	if err := s.db.SelectContext(ctx, &sessions, `SELECT id,provider_session_id FROM platform_checkout_sessions
 WHERE status='active' AND provider_session_id IS NOT NULL ORDER BY id LIMIT $1`, limit); err != nil {
		return result, err
	}
	for _, row := range sessions {
		response, err := s.provider.GetHostedSession(ctx, row.SessionID)
		result.SessionsChecked++
		if err != nil {
			result.Failures++
			continue
		}
		if response.Status != "COMPLETED" && response.Status != "EXPIRED" {
			continue
		}
		var ref string
		if err = s.db.GetContext(ctx, &ref, `SELECT reference_id FROM platform_checkout_sessions WHERE id=$1`, row.ID); err != nil {
			result.Failures++
			continue
		}
		var event sessionEvent
		event.Event = "payment_session.completed"
		if response.Status == "EXPIRED" {
			event.Event = "payment_session.expired"
		}
		event.Data.PaymentSessionID, event.Data.ReferenceID, event.Data.Status = response.PaymentSessionID, ref, response.Status
		event.Data.Currency, event.Data.Amount = response.Currency, response.Amount
		event.Data.PaymentID, event.Data.PaymentTokenID = response.PaymentID, response.PaymentTokenID
		if err = s.applySessionEvent(ctx, event); err != nil {
			result.Failures++
		}
	}
	var charges []struct {
		ID        int64  `db:"id"`
		RequestID string `db:"provider_request_id"`
	}
	if err := s.db.SelectContext(ctx, &charges, `SELECT id,provider_request_id FROM platform_autopay_charges
 WHERE status IN ('pending','review_required') AND provider_request_id IS NOT NULL
 AND next_check_at <= (NOW() AT TIME ZONE 'UTC') ORDER BY id LIMIT $1`, limit); err != nil {
		return result, err
	}
	for _, row := range charges {
		response, err := s.provider.GetRecurringCharge(ctx, row.RequestID)
		result.AutopayChecked++
		if err == nil {
			err = s.applyCharge(ctx, row.ID, response)
		}
		if err != nil {
			result.Failures++
			result.ReviewRequired++
		}
	}
	for i := 0; i < limit; i++ {
		charge, err := s.claimCharge(ctx)
		if err != nil {
			return result, err
		}
		if charge == nil {
			break
		}
		result.AutopayStarted++
		response, err := s.provider.CreateRecurringCharge(ctx, charge.Request)
		if err != nil {
			status, code := "review_required", "charge_outcome_uncertain"
			if isDefinitive(err) {
				status, code = "failed", "provider_rejected"
			}
			if _, saveErr := s.db.ExecContext(context.WithoutCancel(ctx), `UPDATE platform_autopay_charges SET status=$2,failure_code=$3,
 updated_at=(NOW() AT TIME ZONE 'UTC') WHERE id=$1`, charge.ID, status, code); saveErr != nil {
				return result, saveErr
			}
			result.Failures++
			if status == "review_required" {
				result.ReviewRequired++
			}
			continue
		}
		if response == nil || response.PaymentRequestID == "" || response.ReferenceID != charge.Request.ReferenceID {
			result.Failures++
			result.ReviewRequired++
			_, _ = s.db.ExecContext(context.WithoutCancel(ctx), `UPDATE platform_autopay_charges SET status='review_required',failure_code='unverified_response' WHERE id=$1`, charge.ID)
			continue
		}
		if _, err = s.db.ExecContext(context.WithoutCancel(ctx), `UPDATE platform_autopay_charges SET provider_request_id=$2 WHERE id=$1`, charge.ID, response.PaymentRequestID); err != nil {
			return result, err
		}
		if err = s.applyCharge(context.WithoutCancel(ctx), charge.ID, response); err != nil {
			result.Failures++
			result.ReviewRequired++
		}
	}
	return result, nil
}

func (s *Service) renewOne(ctx context.Context) (bool, error) {
	tx, err := s.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	var row struct {
		BusinessID string    `db:"business_id"`
		Anchor     time.Time `db:"billing_anchor"`
		Start      time.Time `db:"paid_through"`
	}
	err = tx.GetContext(ctx, &row, `SELECT business_id,billing_anchor,paid_through FROM platform_subscriptions
 WHERE status='active' AND paid_through <= (NOW() AT TIME ZONE 'UTC')
 AND auto_renew_revoked_at IS NULL ORDER BY paid_through,business_id LIMIT 1 FOR UPDATE SKIP LOCKED`)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	end, err := billing.NextPeriod(row.Anchor, row.Start, "monthly")
	if err != nil {
		return false, err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO platform_invoices(business_id,period_start,period_end,amount,due_date)
 VALUES($1,$2,$3,$4,$2) ON CONFLICT(business_id,period_start) DO NOTHING`, row.BusinessID, row.Start, end, MonthlyPrice); err != nil {
		return false, err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE platform_subscriptions SET status='past_due',next_billing_date=$2,
 updated_at=(NOW() AT TIME ZONE 'UTC') WHERE business_id=$1`, row.BusinessID, end); err != nil {
		return false, err
	}
	return true, tx.Commit()
}

type chargeWork struct {
	ID      int64
	Request provider.RecurringChargeRequest
}

func (s *Service) claimCharge(ctx context.Context) (*chargeWork, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var row struct {
		InvoiceID int64     `db:"invoice_id"`
		Token     string    `db:"payment_token_id"`
		Amount    int64     `db:"amount"`
		Expiry    time.Time `db:"auto_renew_consent_expires_at"`
	}
	err = tx.GetContext(ctx, &row, `SELECT i.id AS invoice_id,s.payment_token_id,i.amount,s.auto_renew_consent_expires_at
 FROM platform_invoices i JOIN platform_subscriptions s ON s.business_id=i.business_id
 WHERE i.status='pending' AND i.due_date <= (NOW() AT TIME ZONE 'UTC') AND s.status='past_due'
 AND s.payment_token_id IS NOT NULL AND s.auto_renew_consent_at IS NOT NULL
 AND s.auto_renew_revoked_at IS NULL AND s.auto_renew_consent_expires_at > (NOW() AT TIME ZONE 'UTC')
 AND i.period_start >= s.auto_renew_consent_at
 AND NOT EXISTS(SELECT 1 FROM platform_autopay_charges ac WHERE ac.invoice_id=i.id)
 AND NOT EXISTS(SELECT 1 FROM platform_checkout_sessions cs WHERE cs.invoice_id=i.id AND cs.status IN ('creating','active'))
 ORDER BY i.due_date,i.id LIMIT 1 FOR UPDATE OF i,s SKIP LOCKED`)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	random := make([]byte, 16)
	if _, err = rand.Read(random); err != nil {
		return nil, err
	}
	reference := "platform_autopay_" + hex.EncodeToString(random)
	var id int64
	if err = tx.GetContext(ctx, &id, `INSERT INTO platform_autopay_charges(invoice_id,reference_id,status)
 VALUES($1,$2,'creating') RETURNING id`, row.InvoiceID, reference); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return &chargeWork{ID: id, Request: provider.RecurringChargeRequest{ReferenceID: reference, PaymentTokenID: row.Token,
		Amount: row.Amount, ReturnURL: s.returnURL + "/dashboard/tagihan-platform", FrequencyDays: 30, ConsentExpiresAt: row.Expiry}}, nil
}

func (s *Service) applyCharge(ctx context.Context, id int64, response *provider.RecurringChargeResponse) error {
	if response == nil || response.PaymentRequestID == "" {
		return fmt.Errorf("missing payment request identity")
	}
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var row struct {
		InvoiceID     int64     `db:"invoice_id"`
		BusinessID    string    `db:"business_id"`
		Reference     string    `db:"reference_id"`
		RequestID     *string   `db:"provider_request_id"`
		Amount        int64     `db:"amount"`
		Status        string    `db:"status"`
		InvoiceStatus string    `db:"invoice_status"`
		PeriodEnd     time.Time `db:"period_end"`
	}
	err = tx.GetContext(ctx, &row, `SELECT ac.invoice_id,i.business_id,ac.reference_id,ac.provider_request_id,
 i.amount,ac.status,i.status AS invoice_status,i.period_end FROM platform_autopay_charges ac
 JOIN platform_invoices i ON i.id=ac.invoice_id WHERE ac.id=$1 FOR UPDATE OF ac,i`, id)
	if err != nil {
		return err
	}
	if row.Reference != response.ReferenceID || row.Amount != response.Amount || response.Currency != "IDR" ||
		(row.RequestID != nil && *row.RequestID != response.PaymentRequestID) {
		return fmt.Errorf("payment request identity mismatch")
	}
	if row.Status == "succeeded" {
		return nil
	}
	status, code := "pending", ""
	switch response.Status {
	case "SUCCEEDED":
		if response.PaymentID == "" || row.InvoiceStatus != "pending" {
			return fmt.Errorf("payment success requires review")
		}
		status = "succeeded"
		if _, err = tx.ExecContext(ctx, `UPDATE platform_invoices SET status='paid',paid_at=(NOW() AT TIME ZONE 'UTC') WHERE id=$1`, row.InvoiceID); err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, `UPDATE platform_subscriptions SET status=CASE WHEN status='cancelled' THEN 'cancelled' ELSE 'active' END,paid_through=$2,
 next_billing_date=$2,updated_at=(NOW() AT TIME ZONE 'UTC') WHERE business_id=$1`, row.BusinessID, row.PeriodEnd); err != nil {
			return err
		}
	case "FAILED", "EXPIRED", "CANCELED":
		status, code = "failed", "provider_declined"
	case "REQUIRES_ACTION":
		status, code = "review_required", "customer_action_required"
	case "ACCEPTING_PAYMENTS", "AUTHORIZED":
	default:
		status, code = "review_required", "unknown_provider_status"
	}
	if _, err = tx.ExecContext(ctx, `UPDATE platform_autopay_charges SET provider_request_id=$2,status=$3,
 failure_code=NULLIF($4,''),next_check_at=(NOW() AT TIME ZONE 'UTC')+INTERVAL '2 minutes',
 updated_at=(NOW() AT TIME ZONE 'UTC') WHERE id=$1`, id, response.PaymentRequestID, status, code); err != nil {
		return err
	}
	return tx.Commit()
}
