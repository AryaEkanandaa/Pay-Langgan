-- +goose Up
CREATE TABLE trial_card_sessions (
    id BIGSERIAL PRIMARY KEY,
    subscription_id INT NOT NULL UNIQUE REFERENCES subscriptions(id),
    public_token TEXT NOT NULL UNIQUE,
    reference_id TEXT NOT NULL UNIQUE,
    provider TEXT NOT NULL CHECK (provider IN ('xendit','midtrans')),
    account_fingerprint TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'created' CHECK (status IN ('created','creating','active','completed','expired')),
    provider_session_id TEXT,
    trial_days INT NOT NULL CHECK (trial_days > 0),
    amount NUMERIC(15,2) NOT NULL,
    items JSONB NOT NULL,
    registration_token TEXT,
    verification_level TEXT CHECK (verification_level IN ('provider_confirmed','registered')),
    consented_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ NOT NULL,
    completed_at TIMESTAMPTZ,
    last_checked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- Saved tokens belong to this gateway account. They are deliberately NOT added
-- to subscription_autopay: registration alone is not a mandate or a payment.

-- +goose Down
DROP TABLE trial_card_sessions;
