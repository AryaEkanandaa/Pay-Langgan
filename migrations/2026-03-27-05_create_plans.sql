-- +goose Up
CREATE TABLE plans (
    id SERIAL PRIMARY KEY,
    product_id INT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    price NUMERIC(12,2) NOT NULL,
    billing_cycle VARCHAR(20) NOT NULL,
    trial_days INT NOT NULL DEFAULT 0,
    meta JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_plans_billing_cycle CHECK (billing_cycle IN ('monthly', 'yearly'))
);

-- +goose Down
DROP TABLE IF EXISTS plans;
