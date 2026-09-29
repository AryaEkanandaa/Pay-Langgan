-- +goose Up
CREATE TABLE checkout_sessions (
    id BIGSERIAL PRIMARY KEY,
    invoice_id INT NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    payment_id INT REFERENCES payments(id) ON DELETE SET NULL,
    public_token_hash CHAR(64) NOT NULL UNIQUE,
    reference_id VARCHAR(255) NOT NULL UNIQUE,
    provider VARCHAR(50) NOT NULL DEFAULT 'xendit',
    provider_session_id VARCHAR(255) UNIQUE,
    payment_link_url TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'created',
    customer_email VARCHAR(254),
    customer_mobile VARCHAR(20),
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    updated_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    CONSTRAINT chk_checkout_sessions_provider CHECK (provider IN ('xendit')),
    CONSTRAINT chk_checkout_sessions_status CHECK (status IN ('created', 'creating', 'active', 'completed', 'expired', 'failed'))
);

CREATE INDEX idx_checkout_sessions_invoice_id ON checkout_sessions(invoice_id);
CREATE INDEX idx_checkout_sessions_expires_at ON checkout_sessions(expires_at);
CREATE UNIQUE INDEX uq_webhook_events_delivery
    ON webhook_events(provider, event_type, reference_id)
    WHERE reference_id IS NOT NULL;
CREATE UNIQUE INDEX uq_payment_methods_provider_token
    ON payment_methods(provider, token);

-- +goose Down
DROP INDEX IF EXISTS uq_payment_methods_provider_token;
DROP INDEX IF EXISTS uq_webhook_events_delivery;
DROP TABLE IF EXISTS checkout_sessions;
