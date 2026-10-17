-- +goose Up
CREATE TABLE checkout_email_deliveries (
 id BIGSERIAL PRIMARY KEY,
 event_type VARCHAR(30) NOT NULL CHECK(event_type IN ('payment_succeeded','trial_started')),
 invoice_id INTEGER REFERENCES invoices(id) ON DELETE CASCADE,
 subscription_id INTEGER NOT NULL REFERENCES subscriptions(id) ON DELETE CASCADE,
 business_id VARCHAR(24) NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
 business_name TEXT NOT NULL,
 customer_name TEXT NOT NULL,
 recipient TEXT NOT NULL,
 settings JSONB NOT NULL DEFAULT '{}',
 snapshot JSONB NOT NULL DEFAULT '{}',
 public_token VARCHAR(64) NOT NULL UNIQUE,
 expires_at TIMESTAMPTZ NOT NULL DEFAULT NOW()+INTERVAL '90 days',
 status VARCHAR(20) NOT NULL DEFAULT 'queued' CHECK(status IN ('queued','sending','sent','previewed','failed')),
 attempts INTEGER NOT NULL DEFAULT 0,
 next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 last_error VARCHAR(60),
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 sent_at TIMESTAMPTZ,
 CHECK ((event_type='payment_succeeded' AND invoice_id IS NOT NULL) OR (event_type='trial_started' AND invoice_id IS NULL)),
 UNIQUE(event_type,invoice_id)
);
CREATE UNIQUE INDEX idx_checkout_trial_email ON checkout_email_deliveries(subscription_id) WHERE event_type='trial_started';
CREATE INDEX idx_checkout_email_queue ON checkout_email_deliveries(next_attempt_at,id) WHERE status IN ('queued','sending');
-- +goose Down
DROP TABLE checkout_email_deliveries;
