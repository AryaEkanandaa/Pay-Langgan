-- +goose Up
ALTER TABLE subscriptions DROP CONSTRAINT chk_subscriptions_status;
ALTER TABLE subscriptions ADD CONSTRAINT chk_subscriptions_status CHECK(status IN ('pending','trial','active','cancelled','paused'));

-- Only correct proven checkout orders that have never had a paid invoice.
-- Existing renewals, paid subscriptions, and admin/legacy records are preserved.
UPDATE subscriptions s SET status='pending'
WHERE s.status='active'
 AND EXISTS(SELECT 1 FROM merchant_orders o JOIN invoices i ON i.id=o.invoice_id WHERE COALESCE(o.subscription_id,i.subscription_id)=s.id AND i.status='pending' AND i.amount>0)
 AND NOT EXISTS(SELECT 1 FROM invoices paid WHERE paid.subscription_id=s.id AND paid.status='paid');

-- +goose Down
-- Preserve disabled access when rolling back; never activate unpaid subscriptions.
UPDATE subscriptions SET status='paused' WHERE status='pending';
ALTER TABLE subscriptions DROP CONSTRAINT chk_subscriptions_status;
ALTER TABLE subscriptions ADD CONSTRAINT chk_subscriptions_status CHECK(status IN ('trial','active','cancelled','paused'));
