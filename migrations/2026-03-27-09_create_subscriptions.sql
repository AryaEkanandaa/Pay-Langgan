-- +goose Up
CREATE TABLE subscriptions (
    id SERIAL PRIMARY KEY,
    customer_id INT NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    plan_id INT NOT NULL REFERENCES plans(id),
    status VARCHAR(20) NOT NULL,
    start_date TIMESTAMP NOT NULL DEFAULT NOW(),
    next_billing_date TIMESTAMP,
    end_date TIMESTAMP,
    trial_ends_at TIMESTAMP,
    meta JSONB NOT NULL DEFAULT '{}'::jsonb,
    CONSTRAINT chk_subscriptions_status CHECK (status IN ('trial', 'active', 'cancelled', 'paused'))
);

-- +goose Down
DROP TABLE IF EXISTS subscriptions;
