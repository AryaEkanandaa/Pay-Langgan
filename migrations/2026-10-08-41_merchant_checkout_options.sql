-- +goose Up
ALTER TABLE merchant_orders ADD COLUMN subscription_id INT REFERENCES subscriptions(id);
ALTER TABLE merchant_orders ADD COLUMN request_hash CHAR(64);
UPDATE merchant_orders o SET subscription_id=i.subscription_id FROM invoices i WHERE i.id=o.invoice_id;
CREATE INDEX idx_merchant_orders_subscription ON merchant_orders(subscription_id);

-- +goose Down
DROP INDEX idx_merchant_orders_subscription;
ALTER TABLE merchant_orders DROP COLUMN request_hash;
ALTER TABLE merchant_orders DROP COLUMN subscription_id;
