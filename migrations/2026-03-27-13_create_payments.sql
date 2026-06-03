-- +goose Up
CREATE TABLE payments (
    id SERIAL PRIMARY KEY,
    invoice_id INT NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,
    provider_transaction_id VARCHAR(255),
    amount NUMERIC(12,2) NOT NULL,
    status VARCHAR(20) NOT NULL,
    paid_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_payments_status CHECK (status IN ('pending', 'success', 'failed', 'expired'))
);

-- +goose Down
DROP TABLE IF EXISTS payments;
