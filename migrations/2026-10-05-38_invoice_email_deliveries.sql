-- +goose Up
CREATE TABLE invoice_email_deliveries (
 id BIGSERIAL PRIMARY KEY,
 invoice_id INTEGER NOT NULL UNIQUE REFERENCES invoices(id) ON DELETE CASCADE,
 business_id VARCHAR(24) NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
 business_name TEXT NOT NULL,
 customer_name TEXT NOT NULL,
 recipient TEXT NOT NULL,
 settings JSONB NOT NULL DEFAULT '{}',
 public_token VARCHAR(64) NOT NULL UNIQUE,
 expires_at TIMESTAMPTZ NOT NULL DEFAULT NOW()+INTERVAL '90 days',
 status VARCHAR(20) NOT NULL DEFAULT 'queued' CHECK(status IN ('queued','sending','sent','previewed','failed')),
 attempts INTEGER NOT NULL DEFAULT 0,
 next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 last_error VARCHAR(60),
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 sent_at TIMESTAMPTZ
);
CREATE INDEX idx_invoice_email_queue ON invoice_email_deliveries(next_attempt_at,id) WHERE status IN ('queued','sending');
CREATE INDEX idx_invoice_email_business ON invoice_email_deliveries(business_id,invoice_id);
-- +goose Down
DROP TABLE invoice_email_deliveries;
