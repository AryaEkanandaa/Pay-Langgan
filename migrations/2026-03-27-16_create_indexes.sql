-- +goose Up
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_business_id ON users(business_id);
CREATE INDEX idx_services_business_id ON services(business_id);
CREATE INDEX idx_customers_business_id ON customers(business_id);
CREATE INDEX idx_subscriptions_customer_id ON subscriptions(customer_id);
CREATE INDEX idx_subscriptions_plan_id ON subscriptions(plan_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);
CREATE INDEX idx_subscriptions_next_billing_date ON subscriptions(next_billing_date);
CREATE INDEX idx_invoices_subscription_id ON invoices(subscription_id);
CREATE INDEX idx_invoices_status ON invoices(status);
CREATE INDEX idx_payments_invoice_id ON payments(invoice_id);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_billing_queue_processed ON billing_queue(processed);
CREATE INDEX idx_billing_queue_scheduled_at ON billing_queue(scheduled_at);
CREATE INDEX idx_billing_queue_subscription_id ON billing_queue(subscription_id);
CREATE INDEX idx_coupons_code ON coupons(code);
CREATE INDEX idx_referral_urls_code ON referral_urls(code);

-- +goose Down
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_business_id;
DROP INDEX IF EXISTS idx_services_business_id;
DROP INDEX IF EXISTS idx_customers_business_id;
DROP INDEX IF EXISTS idx_subscriptions_customer_id;
DROP INDEX IF EXISTS idx_subscriptions_plan_id;
DROP INDEX IF EXISTS idx_subscriptions_status;
DROP INDEX IF EXISTS idx_subscriptions_next_billing_date;
DROP INDEX IF EXISTS idx_invoices_subscription_id;
DROP INDEX IF EXISTS idx_invoices_status;
DROP INDEX IF EXISTS idx_payments_invoice_id;
DROP INDEX IF EXISTS idx_payments_status;
DROP INDEX IF EXISTS idx_billing_queue_processed;
DROP INDEX IF EXISTS idx_billing_queue_scheduled_at;
DROP INDEX IF EXISTS idx_billing_queue_subscription_id;
DROP INDEX IF EXISTS idx_coupons_code;
DROP INDEX IF EXISTS idx_referral_urls_code;
