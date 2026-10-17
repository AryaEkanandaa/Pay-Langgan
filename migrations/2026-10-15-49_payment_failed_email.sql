-- +goose Up
ALTER TABLE checkout_email_deliveries DROP CONSTRAINT checkout_email_deliveries_event_type_check;
ALTER TABLE checkout_email_deliveries DROP CONSTRAINT checkout_email_deliveries_check;
ALTER TABLE checkout_email_deliveries ADD CONSTRAINT checkout_email_deliveries_event_type_check CHECK(event_type IN ('payment_succeeded','payment_failed','trial_started','trial_ending','invoice_due'));
ALTER TABLE checkout_email_deliveries ADD CONSTRAINT checkout_email_deliveries_check CHECK((event_type IN ('payment_succeeded','payment_failed','invoice_due') AND invoice_id IS NOT NULL) OR (event_type IN ('trial_started','trial_ending') AND invoice_id IS NULL));

-- +goose Down
DELETE FROM checkout_email_deliveries WHERE event_type='payment_failed';
ALTER TABLE checkout_email_deliveries DROP CONSTRAINT checkout_email_deliveries_event_type_check;
ALTER TABLE checkout_email_deliveries DROP CONSTRAINT checkout_email_deliveries_check;
ALTER TABLE checkout_email_deliveries ADD CONSTRAINT checkout_email_deliveries_event_type_check CHECK(event_type IN ('payment_succeeded','trial_started','trial_ending','invoice_due'));
ALTER TABLE checkout_email_deliveries ADD CONSTRAINT checkout_email_deliveries_check CHECK((event_type IN ('payment_succeeded','invoice_due') AND invoice_id IS NOT NULL) OR (event_type IN ('trial_started','trial_ending') AND invoice_id IS NULL));
