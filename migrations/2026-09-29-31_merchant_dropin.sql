-- +goose Up
CREATE TABLE merchant_api_keys (
    id BIGSERIAL PRIMARY KEY,
    business_id VARCHAR(50) NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    key_hash CHAR(64) NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    revoked_at TIMESTAMP
);
CREATE INDEX idx_merchant_api_keys_business ON merchant_api_keys(business_id);

CREATE TABLE merchant_orders (
    id BIGSERIAL PRIMARY KEY,
    business_id VARCHAR(50) NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    external_order_id VARCHAR(100) NOT NULL,
    plan_id INT NOT NULL REFERENCES plans(id),
    customer_email VARCHAR(100) NOT NULL,
    invoice_id INT UNIQUE REFERENCES invoices(id),
    created_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    UNIQUE (business_id, external_order_id)
);
CREATE INDEX idx_merchant_orders_business ON merchant_orders(business_id, id DESC);

-- +goose Down
DROP TABLE IF EXISTS merchant_orders;
DROP TABLE IF EXISTS merchant_api_keys;
