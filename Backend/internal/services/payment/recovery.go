package payment

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"pay-langgan/internal/models"
	provider "pay-langgan/pkg/payment"
)

type MaintenanceResult struct {
	AutopayStarted    int `json:"autopay_started"`
	AutopayChecked    int `json:"autopay_checked"`
	WebhooksProcessed int `json:"webhooks_processed"`
	SessionsChecked   int `json:"sessions_checked"`
	ReviewRequired    int `json:"review_required"`
	Failures          int `json:"failures"`
}

// RunMaintenance reserves bounded batches. A lease makes work recoverable after
// process death; state transitions themselves are protected by database locks.
func (s *CheckoutService) RunMaintenance(ctx context.Context, limit int) (MaintenanceResult, error) {
	var result MaintenanceResult
	if limit < 1 || limit > 1000 {
		return result, fmt.Errorf("payment batch limit must be 1..1000")
	}
	result.Failures += s.reconcileOpenTrialCards(ctx, limit)
	if err := s.releaseExpiredTrialCards(ctx, s.db, limit); err != nil {
		return result, err
	}
	if _, err := s.db.ExecContext(ctx, `WITH expired AS (SELECT bucket_key FROM checkout_rate_limits WHERE window_start<(NOW() AT TIME ZONE 'UTC')-INTERVAL '1 hour' ORDER BY window_start LIMIT $1)
 DELETE FROM checkout_rate_limits r USING expired WHERE r.bucket_key=expired.bucket_key`, limit*10); err != nil {
		return result, err
	}
	if _, err := s.db.ExecContext(ctx, `WITH candidate AS (
 SELECT id FROM checkout_sessions WHERE status IN ('created','failed') AND expires_at <= (NOW() AT TIME ZONE 'UTC')
 ORDER BY expires_at,id FOR UPDATE SKIP LOCKED LIMIT $1)
 UPDATE checkout_sessions cs SET status='expired',updated_at=(NOW() AT TIME ZONE 'UTC')
 FROM candidate WHERE cs.id=candidate.id`, limit); err != nil {
		return result, err
	}
	var events []int64
	if err := s.db.SelectContext(ctx, &events, `SELECT id FROM webhook_events WHERE provider='xendit'
 AND event_type IN ('payment_session.completed','payment_session.expired')
 AND processed=FALSE AND dead_letter_at IS NULL AND next_attempt_at <= (NOW() AT TIME ZONE 'UTC')
 ORDER BY next_attempt_at,id LIMIT $1`, limit); err != nil {
		return result, err
	}
	for _, id := range events {
		processed, err := s.ProcessStoredWebhook(ctx, id)
		if err != nil {
			result.Failures++
		}
		if processed {
			result.WebhooksProcessed++
		}
		if ctx.Err() != nil {
			return result, ctx.Err()
		}
	}
	// Reserve just one session at a time so a long batch cannot outlive leases
	// of records it has not started yet. Each provider call has a 15s timeout.
	for i := 0; i < limit; i++ {
		var id int64
		err := s.db.QueryRowContext(ctx, `WITH candidate AS (
 SELECT id FROM checkout_sessions WHERE status IN ('creating','active')
 AND next_reconcile_at <= (NOW() AT TIME ZONE 'UTC')
 AND updated_at <= (NOW() AT TIME ZONE 'UTC') - INTERVAL '30 seconds'
 ORDER BY next_reconcile_at,id FOR UPDATE SKIP LOCKED LIMIT 1)
 UPDATE checkout_sessions cs SET next_reconcile_at=(NOW() AT TIME ZONE 'UTC') + INTERVAL '2 minutes',
 reconcile_attempts=reconcile_attempts+1 FROM candidate WHERE cs.id=candidate.id RETURNING cs.id`).Scan(&id)
		if err == sql.ErrNoRows {
			break
		}
		if err != nil {
			return result, err
		}
		record, err := s.repo.FindByID(ctx, id)
		if err != nil {
			return result, err
		}
		if record == nil {
			continue
		}
		if record.ProviderSessionID == nil {
			result.ReviewRequired++
			if err := s.noteRecovery(ctx, id, "awaiting_session_identity", true, 10*time.Minute); err != nil {
				return result, err
			}
			continue
		}
		result.SessionsChecked++
		if err := s.reconcileRecord(ctx, record, *record.ProviderSessionID); err != nil {
			result.Failures++
			result.ReviewRequired++
			if ctx.Err() != nil {
				return result, ctx.Err()
			}
			if err := s.noteRecovery(ctx, id, "provider_reconciliation_failed", true, 5*time.Minute); err != nil {
				return result, err
			}
		} else if err := s.noteRecovery(ctx, id, "", false, 2*time.Minute); err != nil {
			return result, err
		}
	}
	if err := s.runAutopay(ctx, limit, &result); err != nil {
		return result, err
	}
	return result, nil
}

func (s *CheckoutService) noteRecovery(ctx context.Context, id int64, code string, review bool, delay time.Duration) error {
	_, err := s.db.ExecContext(ctx, `UPDATE checkout_sessions SET last_error=NULLIF($2,''),review_required=$3,next_reconcile_at=$4
 WHERE id=$1 AND status IN ('creating','active')`, id, code, review, time.Now().UTC().Add(delay))
	return err
}

// RecoverCheckout can supply an ID found in the provider dashboard when the
// creation response was lost. The ID, reference, currency and amount are verified
// over the provider API before any local payment status is changed.
func (s *CheckoutService) RecoverCheckout(ctx context.Context, id int64, sessionID string) error {
	record, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if record == nil {
		return fmt.Errorf("checkout not found")
	}
	if record.Status != "creating" && record.Status != "active" {
		return fmt.Errorf("checkout is not recoverable")
	}
	if sessionID == "" && record.ProviderSessionID != nil {
		sessionID = *record.ProviderSessionID
	}
	if sessionID == "" {
		return fmt.Errorf("provider session ID is required; never recreate an uncertain session")
	}
	if err := s.reconcileRecord(ctx, record, sessionID); err != nil {
		return err
	}
	return s.noteRecovery(ctx, id, "", false, 2*time.Minute)
}

func (s *CheckoutService) reconcileRecord(ctx context.Context, record *models.CheckoutRecord, sessionID string) error {
	if record.Provider == "midtrans" {
		if record.CheckoutMode == "CORE" {
			gateway, err := s.gatewayForBusiness(ctx, record.BusinessID)
			if err != nil {
				return err
			}
			if gateway == nil || gateway.Config.Provider != "midtrans" {
				return fmt.Errorf("Midtrans gateway configuration changed")
			}
			transaction, err := s.midtransCardProvider(gateway).GetCardTransaction(ctx, record.ReferenceID)
			if err != nil {
				return err
			}
			return s.acceptMidtransCardTransaction(ctx, record, gateway, transaction)
		}
		if record.ProviderSessionID == nil || *record.ProviderSessionID != sessionID {
			return fmt.Errorf("provider session identity conflict")
		}
		if record.Status == "creating" {
			if err := s.repo.Activate(ctx, record.ID, record.InvoiceID, "midtrans", record.InvoiceAmount, sessionID, record.ExpiresAt); err != nil {
				return err
			}
		}
		// Snap payment outcomes arrive through the signed Midtrans notification.
		// Do not poll Xendit or infer payment status from the browser's callback.
		return nil
	}
	if record.ProviderSessionID != nil && *record.ProviderSessionID != sessionID {
		return fmt.Errorf("provider session identity conflict")
	}
	accountProvider, err := s.xenditForBusiness(ctx, record.BusinessID)
	if err != nil {
		return err
	}
	session, err := accountProvider.GetCardSession(ctx, sessionID, record.ReferenceID)
	if err != nil {
		return err
	}
	if session == nil || session.PaymentSessionID != sessionID || session.ReferenceID != record.ReferenceID ||
		session.Currency != "IDR" || float64(session.Amount) != record.InvoiceAmount {
		return fmt.Errorf("provider session does not match checkout")
	}
	switch session.Status {
	case "ACTIVE":
		if session.ExpiresAt.IsZero() {
			return fmt.Errorf("provider session expiry missing")
		}
		if record.Status == "creating" {
			if session.Mode != "COMPONENTS" {
				return fmt.Errorf("recovered session must use components")
			}
			if err := s.repo.RememberSession(ctx, record.ID, sessionID); err != nil {
				return err
			}
			if err := s.repo.Activate(ctx, record.ID, record.InvoiceID, "xendit", record.InvoiceAmount, sessionID, session.ExpiresAt); err != nil {
				return err
			}
		}
		// Local expiry is not proof of non-payment. Keep blocking new sessions
		// until Xendit confirms a terminal result.
		return nil
	case "COMPLETED", "EXPIRED":
		event := sessionEvent(session)
		tx, err := s.db.BeginTxx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
		if err := s.applySessionEvent(ctx, tx, event); err != nil {
			return err
		}
		return tx.Commit()
	default:
		// CANCELED and future statuses require review until their payment
		// semantics are explicitly supported. Do not infer a failed payment.
		return fmt.Errorf("provider session status requires review")
	}
}

func sessionEvent(session *provider.XenditSessionResponse) xenditSessionWebhook {
	var event xenditSessionWebhook
	event.Event = "payment_session.expired"
	if session.Status == "COMPLETED" {
		event.Event = "payment_session.completed"
	}
	event.Data.PaymentSessionID = session.PaymentSessionID
	event.Data.ReferenceID = session.ReferenceID
	event.Data.Status = session.Status
	event.Data.Currency = session.Currency
	event.Data.Amount = session.Amount
	event.Data.PaymentID = session.PaymentID
	event.Data.PaymentRequestID = session.PaymentRequestID
	event.Data.PaymentTokenID = session.PaymentTokenID
	return event
}
