package payment_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"pay-langgan/internal/config"
	billingrepo "pay-langgan/internal/repositories/billing"
	"pay-langgan/internal/services/billing"
	"pay-langgan/internal/utils"
)

// Preview mode and an isolated schema: never sends email to a real recipient.
func TestEmailHistoryRemindersPostgres(t *testing.T) {
	db := paymentTestDB(t)
	ctx := context.Background()
	svc := billing.NewInvoiceService(db, billingrepo.NewInvoiceRepository(db))
	exec := func(query string, args ...interface{}) {
		t.Helper()
		if _, err := db.Exec(query, args...); err != nil {
			t.Fatal(err)
		}
	}
	count := func(query string) int {
		t.Helper()
		var n int
		if err := db.Get(&n, query); err != nil {
			t.Fatal(err)
		}
		return n
	}
	p, err := svc.ReminderPreferences(ctx, "test")
	if err != nil || p.Enabled {
		t.Fatalf("default opt-out: %+v %v", p, err)
	}
	if err = svc.QueueDueReminders(ctx, time.Now()); err != nil {
		t.Fatal(err)
	}
	if count(`SELECT count(*) FROM checkout_email_deliveries`) != 0 {
		t.Fatal("queued without opt-in")
	}
	if _, err = svc.SetReminderPreferences(ctx, "test", true); err != nil {
		t.Fatal(err)
	}
	exec(`UPDATE email_reminder_preferences SET enabled_at=NOW()-INTERVAL '1 hour';
 UPDATE invoices SET due_date=(NOW() AT TIME ZONE 'UTC')+INTERVAL '12 hours',created_at=NOW() WHERE id=1;
 INSERT INTO subscriptions(id,customer_id,plan_id,status,start_date,billing_cycle,billing_anchor,trial_ends_at,created_at) VALUES
 (2,1,1,'trial',NOW(),'monthly',NOW(),(NOW() AT TIME ZONE 'UTC')+INTERVAL '2 days',NOW()),
 (3,1,1,'trial',NOW(),'monthly',NOW(),(NOW() AT TIME ZONE 'UTC')+INTERVAL '2 days',NOW()-INTERVAL '2 hours'),
 (4,1,1,'trial',NOW(),'monthly',NOW(),(NOW() AT TIME ZONE 'UTC')+INTERVAL '5 days',NOW()),
 (5,1,1,'cancelled',NOW(),'monthly',NOW(),(NOW() AT TIME ZONE 'UTC')+INTERVAL '2 days',NOW());
 INSERT INTO invoices(id,subscription_id,invoice_number,amount,status,due_date,created_at) VALUES
 (2,1,'INV-OLD',100,'pending',NOW()+INTERVAL '12 hours',NOW()-INTERVAL '2 hours'),
 (3,1,'INV-PAID',100,'paid',NOW()+INTERVAL '12 hours',NOW()),
 (4,1,'INV-FUTURE',100,'pending',NOW()+INTERVAL '3 days',NOW());
 INSERT INTO businesses(id,name) VALUES('other','Other');`)
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if e := svc.QueueDueReminders(ctx, time.Now()); e != nil {
				t.Error(e)
			}
		}()
	}
	wg.Wait()
	if count(`SELECT count(*) FROM checkout_email_deliveries`) != 2 {
		t.Fatal("duplicates or ineligible reminders queued")
	}
	items, total, err := svc.CustomerEmailHistory(ctx, "test", "", "", "", 1)
	if err != nil || total != 2 || len(items) != 2 {
		t.Fatalf("history: %d %+v %v", total, items, err)
	}
	_, total, err = svc.CustomerEmailHistory(ctx, "other", "", "", "", 1)
	if err != nil || total != 0 {
		t.Fatal("history leaked to another tenant", err)
	}
	_, total, err = svc.CustomerEmailHistory(ctx, "test", "queued", "trial_ending", "Alya", 1)
	if err != nil || total != 1 {
		t.Fatal("filters", total, err)
	}
	if _, _, err = svc.CustomerEmailHistory(ctx, "test", "invalid", "", "", 1); !errors.Is(err, utils.ErrBadRequest) {
		t.Fatal("invalid filter accepted", err)
	}
	var trialID, invoiceID int
	if err = db.Get(&trialID, `SELECT id FROM checkout_email_deliveries WHERE event_type='trial_ending'`); err != nil {
		t.Fatal(err)
	}
	if err = db.Get(&invoiceID, `SELECT id FROM checkout_email_deliveries WHERE event_type='invoice_due'`); err != nil {
		t.Fatal(err)
	}
	if _, _, err = svc.CustomerEmailPreview(ctx, "test", "checkout", trialID); !errors.Is(err, utils.ErrConflict) {
		t.Fatal("invented unprocessed preview", err)
	}
	cfg := config.Load()
	cfg.EmailDeliveryMode = "preview"
	cfg.EmailPreviewDir = t.TempDir()
	cfg.APIPublicURL = "https://backend.example.invalid"
	if err = svc.RunCheckoutEmailBatch(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	if count(`SELECT count(*) FROM checkout_email_deliveries WHERE status='previewed' AND sent_at IS NULL AND rendered_html IS NOT NULL`) != 2 {
		t.Fatal("preview claimed delivery or missing rendered HTML")
	}
	subject, html, err := svc.CustomerEmailPreview(ctx, "test", "checkout", trialID)
	if err != nil || !strings.Contains(subject, "Trial") || !strings.Contains(html, "Trial segera berakhir") {
		t.Fatal("trial preview", subject, err)
	}
	if _, _, err = svc.CustomerEmailPreview(ctx, "other", "checkout", trialID); !errors.Is(err, utils.ErrNotFound) {
		t.Fatal("preview leaked", err)
	}
	var token string
	if err = db.Get(&token, `SELECT public_token FROM checkout_email_deliveries WHERE id=$1`, invoiceID); err != nil {
		t.Fatal(err)
	}
	id, business, _, err := svc.PublicEmailInvoice(ctx, token)
	if err != nil || id != 1 || business != "test" {
		t.Fatal("reminder invoice link", id, business, err)
	}
	// Only failed emails can be retried; a simultaneous second retry must lose.
	if err = svc.RetryCustomerEmail(ctx, "test", "checkout", trialID); !errors.Is(err, utils.ErrConflict) {
		t.Fatal("retried successful preview", err)
	}
	exec(`UPDATE checkout_email_deliveries SET status='failed' WHERE id=$1`, trialID)
	if err = svc.RetryCustomerEmail(ctx, "other", "checkout", trialID); !errors.Is(err, utils.ErrConflict) {
		t.Fatal("tenant retry allowed", err)
	}
	if err = svc.RetryCustomerEmail(ctx, "test", "checkout", trialID); err != nil {
		t.Fatal(err)
	}
	if err = svc.RetryCustomerEmail(ctx, "test", "checkout", trialID); !errors.Is(err, utils.ErrConflict) {
		t.Fatal("duplicate retry allowed", err)
	}
	if _, err = svc.SetReminderPreferences(ctx, "test", false); err != nil {
		t.Fatal(err)
	}
	if err = svc.RunCheckoutEmailBatch(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	if count(`SELECT count(*) FROM checkout_email_deliveries WHERE status='cancelled'`) != 1 {
		t.Fatal("opt-out not honored")
	}
	// Re-check payment and trial state immediately before dispatch.
	exec(`UPDATE email_reminder_preferences SET enabled=TRUE,enabled_at=NOW()-INTERVAL '1 hour';
 UPDATE checkout_email_deliveries SET status='queued',next_attempt_at=NOW(),attempts=0;
 UPDATE invoices SET status='paid' WHERE id=1;
 UPDATE subscriptions SET status='cancelled' WHERE id=2;`)
	if err = svc.RunCheckoutEmailBatch(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	if count(`SELECT count(*) FROM checkout_email_deliveries WHERE status='cancelled'`) != 2 {
		t.Fatal("obsolete reminder dispatched")
	}
	// Invoice-issued mail shares the history, but stays a separate event.
	if _, err = svc.QueueEmail(ctx, 1, "test"); err != nil {
		t.Fatal(err)
	}
	if err = svc.RunEmailBatch(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	_, total, err = svc.CustomerEmailHistory(ctx, "test", "previewed", "invoice_issued", "", 1)
	if err != nil || total != 1 {
		t.Fatal("invoice queue missing from history", total, err)
	}
	exec(`INSERT INTO customers(id,business_id,name,email) VALUES(2,'test','Invalid','bad-address');
 INSERT INTO subscriptions(id,customer_id,plan_id,status,start_date,trial_ends_at) VALUES(6,2,1,'trial',NOW(),(NOW() AT TIME ZONE 'UTC')+INTERVAL '2 days');`)
	if err = svc.QueueDueReminders(ctx, time.Now()); err != nil {
		t.Fatal(err)
	}
	if count(`SELECT count(*) FROM checkout_email_deliveries WHERE subscription_id=6 AND status='failed' AND last_error='invalid_recipient'`) != 1 {
		t.Fatal("invalid recipient starves subsequent reminder batches")
	}
}
