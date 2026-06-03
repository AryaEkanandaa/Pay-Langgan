-- +goose Up
CREATE TABLE payment_methods (
    id SERIAL PRIMARY KEY,
    customer_id INT NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,
    token VARCHAR(255) NOT NULL,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_payment_methods_provider CHECK (provider IN ('midtrans', 'xendit', 'doku'))
);

-- +goose Down
DROP TABLE IF EXISTS payment_methods;
