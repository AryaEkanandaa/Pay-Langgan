-- +goose Up
CREATE TABLE customers (
    id SERIAL PRIMARY KEY,
    business_id VARCHAR(50) NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100),
    contact VARCHAR(20),
    meta JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS customers;
