-- +goose Up
CREATE TABLE coupons (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE,
    discount_type VARCHAR(20) NOT NULL,
    discount_value NUMERIC(12,2) NOT NULL,
    max_usage INT,
    used_count INT NOT NULL DEFAULT 0,
    expires_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_coupons_discount_type CHECK (discount_type IN ('percentage', 'fixed'))
);

-- +goose Down
DROP TABLE IF EXISTS coupons;
