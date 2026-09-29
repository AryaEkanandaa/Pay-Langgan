-- +goose Up
CREATE TABLE checkout_rate_limits (
    bucket_key CHAR(64) PRIMARY KEY,
    window_start TIMESTAMP NOT NULL,
    requests INT NOT NULL CHECK (requests > 0)
);
CREATE INDEX idx_checkout_rate_limits_expiry ON checkout_rate_limits(window_start);

CREATE TABLE subscription_autopay (
    subscription_id INT PRIMARY KEY REFERENCES subscriptions(id) ON DELETE CASCADE,
    payment_method_id INT NOT NULL REFERENCES payment_methods(id),
    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    consent_reference VARCHAR(255) NOT NULL,
    consented_at TIMESTAMP NOT NULL,
    consent_expires_at TIMESTAMP NOT NULL,
    max_amount BIGINT NOT NULL CHECK (max_amount BETWEEN 5000 AND 200000000),
    effective_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    updated_by INT NOT NULL REFERENCES users(id),
    updated_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    CHECK (consent_expires_at > consented_at)
);

CREATE TABLE autopay_charges (
    id BIGSERIAL PRIMARY KEY,
    invoice_id INT NOT NULL UNIQUE REFERENCES invoices(id),
    payment_id INT NOT NULL UNIQUE REFERENCES payments(id),
    payment_method_id INT NOT NULL REFERENCES payment_methods(id),
    reference_id VARCHAR(64) NOT NULL UNIQUE,
    provider_request_id VARCHAR(255) UNIQUE,
    status VARCHAR(30) NOT NULL CHECK (status IN ('creating','pending','succeeded','failed','review_required')),
    failure_code VARCHAR(100),
    next_check_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    created_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    updated_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC')
);
CREATE INDEX idx_autopay_due ON autopay_charges(next_check_at,id)
    WHERE status IN ('creating','pending','review_required');

-- +goose Down
DROP TABLE autopay_charges;
DROP TABLE subscription_autopay;
DROP TABLE checkout_rate_limits;
