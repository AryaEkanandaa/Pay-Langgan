package billing

import (
	"context"
	"github.com/jmoiron/sqlx"
	"time"
)

type EmailReminderPreferences struct {
	Enabled   bool       `db:"enabled" json:"enabled"`
	EnabledAt *time.Time `db:"enabled_at" json:"enabled_at,omitempty"`
}

func (s *InvoiceService) ReminderPreferences(ctx context.Context, business string) (EmailReminderPreferences, error) {
	var p EmailReminderPreferences
	err := s.db.GetContext(ctx, &p, `SELECT COALESCE(p.enabled,FALSE) enabled,p.enabled_at FROM businesses b LEFT JOIN email_reminder_preferences p ON p.business_id=b.id WHERE b.id=$1`, business)
	return p, err
}
func (s *InvoiceService) SetReminderPreferences(ctx context.Context, business string, enabled bool) (EmailReminderPreferences, error) {
	_, err := s.db.ExecContext(ctx, `INSERT INTO email_reminder_preferences(business_id,enabled) VALUES($1,$2) ON CONFLICT(business_id) DO UPDATE SET enabled=EXCLUDED.enabled,enabled_at=CASE WHEN EXCLUDED.enabled AND NOT email_reminder_preferences.enabled THEN NOW() ELSE email_reminder_preferences.enabled_at END`, business, enabled)
	if err != nil {
		return EmailReminderPreferences{}, err
	}
	return s.ReminderPreferences(ctx, business)
}

// Only new records since opt-in are eligible, and each kind is uniquely queued.
func (s *InvoiceService) QueueDueReminders(ctx context.Context, at time.Time) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var trials []int
	err = tx.SelectContext(ctx, &trials, `SELECT s.id FROM subscriptions s JOIN customers c ON c.id=s.customer_id JOIN businesses b ON b.id=c.business_id JOIN email_reminder_preferences p ON p.business_id=b.id WHERE p.enabled AND b.status='active' AND NOT b.deleted AND s.status='trial' AND s.created_at>=p.enabled_at::timestamp AND s.trial_ends_at>$1 AND s.trial_ends_at<=$2 AND NOT EXISTS(SELECT 1 FROM checkout_email_deliveries e WHERE e.subscription_id=s.id AND e.event_type='trial_ending') ORDER BY s.trial_ends_at LIMIT 100`, at.UTC(), at.UTC().Add(72*time.Hour))
	if err != nil {
		return err
	}
	for _, id := range trials {
		if err = queueTrialReminder(ctx, tx, id); err != nil {
			return err
		}
	}
	var invoices []struct {
		ID             int `db:"id"`
		SubscriptionID int `db:"subscription_id"`
	}
	err = tx.SelectContext(ctx, &invoices, `SELECT i.id,i.subscription_id FROM invoices i JOIN subscriptions s ON s.id=i.subscription_id JOIN customers c ON c.id=s.customer_id JOIN businesses b ON b.id=c.business_id JOIN email_reminder_preferences p ON p.business_id=b.id WHERE p.enabled AND b.status='active' AND NOT b.deleted AND s.status IN ('pending','active') AND i.status='pending' AND i.amount>0 AND i.created_at>=p.enabled_at::timestamp AND i.due_date>$1 AND i.due_date<=$2 AND NOT EXISTS(SELECT 1 FROM checkout_email_deliveries e WHERE e.invoice_id=i.id AND e.event_type='invoice_due') ORDER BY i.due_date LIMIT 100`, at.UTC(), at.UTC().Add(24*time.Hour))
	if err != nil {
		return err
	}
	for _, i := range invoices {
		if err = queueCheckoutEmail(ctx, tx, "invoice_due", i.SubscriptionID, &i.ID, nil); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func queueTrialReminder(ctx context.Context, tx *sqlx.Tx, id int) error {
	// No invented post-trial price; use the current plan only for the reminder title.
	return queueCheckoutEmail(ctx, tx, "trial_ending", id, nil, nil)
}

func (s *InvoiceService) reminderEligible(ctx context.Context, id int) (bool, error) {
	var eligible bool
	err := s.db.GetContext(ctx, &eligible, `SELECT COALESCE(COALESCE(p.enabled,FALSE) AND b.status='active' AND NOT b.deleted AND
 CASE WHEN e.event_type='trial_ending' THEN s.status='trial' AND s.created_at>=p.enabled_at::timestamp AND s.trial_ends_at>(NOW() AT TIME ZONE 'UTC')
 WHEN e.event_type='invoice_due' THEN s.status IN ('pending','active') AND i.status='pending' AND i.due_date>(NOW() AT TIME ZONE 'UTC') AND i.created_at>=p.enabled_at::timestamp ELSE TRUE END,FALSE)
 FROM checkout_email_deliveries e JOIN subscriptions s ON s.id=e.subscription_id JOIN businesses b ON b.id=e.business_id LEFT JOIN invoices i ON i.id=e.invoice_id LEFT JOIN email_reminder_preferences p ON p.business_id=b.id WHERE e.id=$1`, id)
	return eligible, err
}
