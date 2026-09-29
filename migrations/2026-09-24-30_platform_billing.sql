-- +goose Up
-- PayLanggan's own subscription is separate from each tenant's customer billing.
CREATE TABLE platform_subscriptions (
    business_id VARCHAR(50) PRIMARY KEY REFERENCES businesses(id),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','active','past_due','cancelled')),
    billing_anchor TIMESTAMP,
    next_billing_date TIMESTAMP,
    paid_through TIMESTAMP,
    payment_token_id VARCHAR(255),
    auto_renew_consent_at TIMESTAMP,
    auto_renew_consent_expires_at TIMESTAMP,
    auto_renew_revoked_at TIMESTAMP,
    consent_reference VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    updated_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC')
);

CREATE TABLE platform_invoices (
    id BIGSERIAL PRIMARY KEY,
    business_id VARCHAR(50) NOT NULL REFERENCES businesses(id),
    period_start TIMESTAMP NOT NULL,
    period_end TIMESTAMP NOT NULL,
    amount BIGINT NOT NULL DEFAULT 300000 CHECK (amount = 300000),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','paid')),
    due_date TIMESTAMP NOT NULL,
    paid_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    UNIQUE (business_id, period_start),
    CHECK (period_end > period_start)
);
CREATE INDEX idx_platform_invoices_business ON platform_invoices(business_id,id DESC);

CREATE TABLE platform_checkout_sessions (
    id BIGSERIAL PRIMARY KEY,
    invoice_id BIGINT NOT NULL REFERENCES platform_invoices(id),
    reference_id VARCHAR(100) NOT NULL UNIQUE,
    provider_session_id VARCHAR(255) UNIQUE,
    payment_link_url TEXT,
    status VARCHAR(20) NOT NULL CHECK (status IN ('creating','active','completed','expired','failed')),
    last_error VARCHAR(100),
    created_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    updated_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC')
);
CREATE UNIQUE INDEX idx_platform_checkout_open ON platform_checkout_sessions(invoice_id)
    WHERE status IN ('creating','active');

CREATE TABLE platform_autopay_charges (
    id BIGSERIAL PRIMARY KEY,
    invoice_id BIGINT NOT NULL UNIQUE REFERENCES platform_invoices(id),
    reference_id VARCHAR(100) NOT NULL UNIQUE,
    provider_request_id VARCHAR(255) UNIQUE,
    status VARCHAR(30) NOT NULL CHECK (status IN ('creating','pending','succeeded','failed','review_required')),
    failure_code VARCHAR(100),
    next_check_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    created_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    updated_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC')
);
CREATE INDEX idx_platform_autopay_due ON platform_autopay_charges(next_check_at,id)
    WHERE status IN ('creating','pending','review_required');

-- +goose Down
DROP TABLE platform_autopay_charges;
DROP TABLE platform_checkout_sessions;
DROP TABLE platform_invoices;
DROP TABLE platform_subscriptions;
