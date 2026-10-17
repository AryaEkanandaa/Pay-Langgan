-- +goose Up
-- NOW() backfills existing users at migration time; historical logs stay read.
ALTER TABLE users ADD COLUMN webhook_logs_last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
-- idx_business_webhook_logs_recent already indexes (business_id, received_at, id).

-- +goose Down
ALTER TABLE users DROP COLUMN webhook_logs_last_seen_at;
