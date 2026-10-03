-- +goose Up
CREATE TABLE business_payment_gateways (
    business_id VARCHAR(24) PRIMARY KEY REFERENCES businesses(id) ON DELETE CASCADE,
    provider VARCHAR(20) NOT NULL CHECK (provider IN ('xendit', 'midtrans')),
    environment VARCHAR(12) NOT NULL CHECK (environment IN ('sandbox', 'production')),
    encrypted_credentials BYTEA NOT NULL,
    webhook_id VARCHAR(64) NOT NULL UNIQUE,
    configured_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    verified_at TIMESTAMP WITHOUT TIME ZONE
);

-- +goose Down
DROP TABLE business_payment_gateways;
