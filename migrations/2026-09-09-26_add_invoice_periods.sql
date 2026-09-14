-- +goose Up
ALTER TABLE subscriptions
    ADD COLUMN billing_anchor TIMESTAMP,
    ADD COLUMN billing_cycle VARCHAR(20),
    ADD CONSTRAINT chk_subscription_billing_schedule CHECK (
        (billing_anchor IS NULL AND billing_cycle IS NULL) OR
        (billing_anchor IS NOT NULL AND billing_cycle IS NOT NULL AND billing_cycle IN ('monthly', 'yearly'))
    );

ALTER TABLE invoices
    ADD COLUMN period_start TIMESTAMP,
    ADD COLUMN period_end TIMESTAMP,
    ADD CONSTRAINT chk_invoice_period CHECK (
        (period_start IS NULL AND period_end IS NULL) OR
        (period_start IS NOT NULL AND period_end IS NOT NULL AND period_end > period_start)
    );
CREATE UNIQUE INDEX uq_invoice_subscription_period ON invoices (subscription_id, period_start)
    WHERE period_start IS NOT NULL;
CREATE INDEX idx_subscriptions_due_billing ON subscriptions (next_billing_date, id)
    WHERE status IN ('active', 'trial') AND billing_anchor IS NOT NULL;

-- Legacy schedules remain NULL intentionally: do not guess already billed periods.

-- +goose Down
DROP INDEX idx_subscriptions_due_billing;
DROP INDEX uq_invoice_subscription_period;
ALTER TABLE invoices DROP CONSTRAINT chk_invoice_period, DROP COLUMN period_start, DROP COLUMN period_end;
ALTER TABLE subscriptions DROP CONSTRAINT chk_subscription_billing_schedule, DROP COLUMN billing_anchor, DROP COLUMN billing_cycle;
