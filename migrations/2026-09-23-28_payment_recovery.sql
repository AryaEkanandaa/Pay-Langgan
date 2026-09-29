-- +goose Up
ALTER TABLE checkout_sessions
    ADD COLUMN next_reconcile_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    ADD COLUMN reconcile_attempts INT NOT NULL DEFAULT 0,
    ADD COLUMN last_error TEXT,
    ADD COLUMN review_required BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX idx_checkout_reconcile ON checkout_sessions(next_reconcile_at, id)
    WHERE status IN ('creating', 'active');

ALTER TABLE webhook_events
    ADD COLUMN attempts INT NOT NULL DEFAULT 0,
    ADD COLUMN next_attempt_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    ADD COLUMN last_error TEXT,
    ADD COLUMN dead_letter_at TIMESTAMP;

CREATE INDEX idx_webhook_retry ON webhook_events(next_attempt_at, id)
    WHERE provider = 'xendit' AND processed = FALSE AND dead_letter_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_webhook_retry;
ALTER TABLE webhook_events DROP COLUMN dead_letter_at, DROP COLUMN last_error,
    DROP COLUMN next_attempt_at, DROP COLUMN attempts;
DROP INDEX IF EXISTS idx_checkout_reconcile;
ALTER TABLE checkout_sessions DROP COLUMN review_required, DROP COLUMN last_error,
    DROP COLUMN reconcile_attempts, DROP COLUMN next_reconcile_at;
