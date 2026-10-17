-- +goose Up
CREATE TABLE email_reminder_preferences (
 business_id VARCHAR(24) PRIMARY KEY REFERENCES businesses(id) ON DELETE CASCADE,
 enabled BOOLEAN NOT NULL DEFAULT FALSE,
 enabled_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
ALTER TABLE checkout_email_deliveries DROP CONSTRAINT checkout_email_deliveries_event_type_check;
ALTER TABLE checkout_email_deliveries DROP CONSTRAINT checkout_email_deliveries_check;
ALTER TABLE checkout_email_deliveries DROP CONSTRAINT checkout_email_deliveries_status_check;
ALTER TABLE checkout_email_deliveries ADD CONSTRAINT checkout_email_deliveries_event_type_check CHECK(event_type IN ('payment_succeeded','trial_started','trial_ending','invoice_due'));
ALTER TABLE checkout_email_deliveries ADD CONSTRAINT checkout_email_deliveries_check CHECK((event_type IN ('payment_succeeded','invoice_due') AND invoice_id IS NOT NULL) OR (event_type IN ('trial_started','trial_ending') AND invoice_id IS NULL));
ALTER TABLE checkout_email_deliveries ADD CONSTRAINT checkout_email_deliveries_status_check CHECK(status IN ('queued','sending','sent','previewed','failed','cancelled'));
CREATE UNIQUE INDEX idx_trial_ending_email ON checkout_email_deliveries(subscription_id) WHERE event_type='trial_ending';
ALTER TABLE invoice_email_deliveries ADD COLUMN rendered_subject TEXT, ADD COLUMN rendered_html TEXT;
ALTER TABLE checkout_email_deliveries ADD COLUMN rendered_subject TEXT, ADD COLUMN rendered_html TEXT;
-- +goose Down
DELETE FROM checkout_email_deliveries WHERE event_type IN ('trial_ending','invoice_due');
UPDATE checkout_email_deliveries SET status='failed' WHERE status='cancelled';
DROP INDEX idx_trial_ending_email;
ALTER TABLE checkout_email_deliveries DROP CONSTRAINT checkout_email_deliveries_event_type_check;
ALTER TABLE checkout_email_deliveries DROP CONSTRAINT checkout_email_deliveries_check;
ALTER TABLE checkout_email_deliveries DROP CONSTRAINT checkout_email_deliveries_status_check;
ALTER TABLE checkout_email_deliveries ADD CONSTRAINT checkout_email_deliveries_event_type_check CHECK(event_type IN ('payment_succeeded','trial_started'));
ALTER TABLE checkout_email_deliveries ADD CONSTRAINT checkout_email_deliveries_check CHECK((event_type='payment_succeeded' AND invoice_id IS NOT NULL) OR (event_type='trial_started' AND invoice_id IS NULL));
ALTER TABLE checkout_email_deliveries ADD CONSTRAINT checkout_email_deliveries_status_check CHECK(status IN ('queued','sending','sent','previewed','failed'));
ALTER TABLE invoice_email_deliveries DROP COLUMN rendered_subject, DROP COLUMN rendered_html;
ALTER TABLE checkout_email_deliveries DROP COLUMN rendered_subject, DROP COLUMN rendered_html;
DROP TABLE email_reminder_preferences;
