package payment

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"pay-langgan/internal/utils"
)

const maxWebhookAttempts = 8

// Persist independently of payment updates. Only retain the normalized fields
// needed for processing, not arbitrary payloads, card data or SDK keys.
func (s *CheckoutService) acceptSessionEvent(ctx context.Context, event xenditSessionWebhook) (bool, error) {
	payload, err := json.Marshal(event)
	if err != nil {
		return false, err
	}
	var id int64
	err = s.db.QueryRowContext(ctx, `INSERT INTO webhook_events
 (provider,event_type,reference_id,payload,received_at,created_at)
 VALUES('xendit',$1,$2,$3::jsonb,(NOW() AT TIME ZONE 'UTC'),(NOW() AT TIME ZONE 'UTC'))
 ON CONFLICT (provider,event_type,reference_id) WHERE reference_id IS NOT NULL DO NOTHING RETURNING id`,
		event.Event, event.Data.PaymentSessionID, string(payload)).Scan(&id)
	if err == sql.ErrNoRows {
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("persist webhook inbox: %w", err)
	}
	// Best effort fast path; the durable inbox and worker own further retries.
	// Even if the request disconnects or processing fails, receipt is acknowledged.
	processCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	_, _ = s.ProcessStoredWebhook(processCtx, id)
	return false, nil
}

func retryDelay(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	if attempt > 8 {
		attempt = 8
	}
	delay := 30 * time.Second * time.Duration(1<<uint(attempt-1))
	if delay > time.Hour {
		return time.Hour
	}
	return delay
}

// ProcessStoredWebhook is safe across workers and duplicate deliveries. Payment
// effects and processed=true commit together; failures retain inbox diagnostics.
func (s *CheckoutService) ProcessStoredWebhook(ctx context.Context, id int64) (bool, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	var row struct {
		Payload  []byte `db:"payload"`
		Attempts int    `db:"attempts"`
	}
	err = tx.GetContext(ctx, &row, `SELECT payload,attempts FROM webhook_events WHERE id=$1 AND provider='xendit'
 AND event_type IN ('payment_session.completed','payment_session.expired')
 AND processed=FALSE AND dead_letter_at IS NULL AND next_attempt_at <= (NOW() AT TIME ZONE 'UTC')
 FOR UPDATE SKIP LOCKED`, id)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if _, err = tx.ExecContext(ctx, `SAVEPOINT payment_effects`); err != nil {
		return false, err
	}
	var event xenditSessionWebhook
	processErr := json.Unmarshal(row.Payload, &event)
	if processErr == nil {
		processErr = s.applySessionEvent(ctx, tx, event)
	}
	if processErr == nil {
		_, err = tx.ExecContext(ctx, `UPDATE webhook_events SET processed=TRUE,processed_at=(NOW() AT TIME ZONE 'UTC'),
 attempts=attempts+1,last_error=NULL,dead_letter_at=NULL WHERE id=$1`, id)
	} else {
		if _, err = tx.ExecContext(ctx, `ROLLBACK TO SAVEPOINT payment_effects`); err != nil {
			return false, err
		}
		code := "processing_failed"
		permanent := errors.Is(processErr, utils.ErrBadRequest)
		if permanent {
			code = "event_validation_failed"
		}
		dead := permanent || row.Attempts+1 >= maxWebhookAttempts
		_, err = tx.ExecContext(ctx, `UPDATE webhook_events SET attempts=attempts+1,last_error=$2,
 next_attempt_at=$3,dead_letter_at=CASE WHEN $4 THEN (NOW() AT TIME ZONE 'UTC') ELSE NULL END WHERE id=$1`,
			id, code, time.Now().UTC().Add(retryDelay(row.Attempts+1)), dead)
	}
	if err != nil {
		return false, err
	}
	if err = tx.Commit(); err != nil {
		return false, err
	}
	if processErr != nil {
		return false, fmt.Errorf("webhook %d processing deferred or requires review", id)
	}
	return true, nil
}

// RetryWebhook is an operator action (CLI), never exposed on the public route.
func (s *CheckoutService) RetryWebhook(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, `UPDATE webhook_events SET attempts=0,dead_letter_at=NULL,
 next_attempt_at=(NOW() AT TIME ZONE 'UTC'),last_error=NULL
 WHERE id=$1 AND provider='xendit' AND processed=FALSE
 AND event_type IN ('payment_session.completed','payment_session.expired')`, id)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return fmt.Errorf("webhook is missing, unsupported, or already processed")
	}
	_, err = s.ProcessStoredWebhook(ctx, id)
	return err
}
