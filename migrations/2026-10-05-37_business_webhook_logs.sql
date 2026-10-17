-- +goose Up
CREATE TABLE business_webhook_logs (
    id BIGSERIAL PRIMARY KEY,
    business_id VARCHAR(24) NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    provider VARCHAR(20) NOT NULL,
    event_type VARCHAR(120) NOT NULL DEFAULT '',
    reference_id VARCHAR(160) NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL CHECK (status IN ('accepted','duplicate','failed')),
    http_status INTEGER NOT NULL,
    error_code VARCHAR(40) NOT NULL DEFAULT '',
    duration_ms INTEGER NOT NULL DEFAULT 0,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_business_webhook_logs_recent ON business_webhook_logs(business_id, received_at DESC, id DESC);
-- Only events already tied to a tenant checkout can be included in history.
INSERT INTO business_webhook_logs(business_id,provider,event_type,reference_id,status,http_status,received_at)
SELECT DISTINCT c.business_id,w.provider,w.event_type,cs.reference_id,'accepted',200,w.received_at AT TIME ZONE 'UTC'
FROM webhook_events w JOIN checkout_sessions cs ON cs.provider=w.provider AND (cs.reference_id=w.reference_id OR cs.provider_session_id=w.reference_id)
JOIN invoices i ON i.id=cs.invoice_id JOIN subscriptions s ON s.id=i.subscription_id
JOIN customers c ON c.id=s.customer_id;

-- +goose Down
DROP TABLE business_webhook_logs;
