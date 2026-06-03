-- +goose Up
CREATE TABLE webhook_events (
    id SERIAL PRIMARY KEY,
    provider VARCHAR(50) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    reference_id VARCHAR(255),
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    processed BOOLEAN NOT NULL DEFAULT FALSE,
    received_at TIMESTAMP NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_webhook_events_provider CHECK (provider IN ('midtrans', 'xendit', 'doku'))
);

-- +goose Down
DROP TABLE IF EXISTS webhook_events;
