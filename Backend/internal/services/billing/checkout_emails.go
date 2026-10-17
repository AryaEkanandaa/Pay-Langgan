package billing

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/mail"
	"time"

	"github.com/jmoiron/sqlx"
	"pay-langgan/internal/models"
)

type trialEmailSnapshot struct {
	Plan             string                           `json:"plan"`
	Cycle            string                           `json:"cycle"`
	EndsAt           time.Time                        `json:"ends_at"`
	Amount           float64                          `json:"amount"`
	Items            []models.SubscriptionPreviewItem `json:"items"`
	AutoDebitEnabled bool                             `json:"auto_debit_enabled"`
	PaymentMethod    string                           `json:"payment_method,omitempty"`
}

// QueuePaymentConfirmation must run in the same transaction as verified payment.
// A distinct key from invoice-issued email permits exactly one receipt per invoice.
func QueuePaymentConfirmation(ctx context.Context, tx *sqlx.Tx, invoiceID int) error {
	var subscriptionID int
	if err := tx.GetContext(ctx, &subscriptionID, `SELECT subscription_id FROM invoices WHERE id=$1 AND status='paid'`, invoiceID); err != nil {
		return err
	}
	return queueCheckoutEmail(ctx, tx, "payment_succeeded", subscriptionID, &invoiceID, nil)
}

// A definitive failed charge keeps the invoice pending and sends one
// actionable notice per invoice. A later successful charge supersedes it.
func QueuePaymentFailure(ctx context.Context, tx *sqlx.Tx, invoiceID int, code string) error {
	var subscriptionID int
	if err := tx.GetContext(ctx, &subscriptionID, `SELECT subscription_id FROM invoices WHERE id=$1 AND status='pending'`, invoiceID); err != nil {
		return err
	}
	snapshot, err := json.Marshal(map[string]string{"failure_code": code})
	if err != nil {
		return err
	}
	return queueCheckoutEmail(ctx, tx, "payment_failed", subscriptionID, &invoiceID, snapshot)
}

func QueueTrialConfirmation(ctx context.Context, tx *sqlx.Tx, subscriptionID int, amount float64, items []models.SubscriptionPreviewItem) error {
	var snapshot trialEmailSnapshot
	if err := tx.QueryRowContext(ctx, `SELECT p.name,s.billing_cycle,s.trial_ends_at,COALESCE(a.enabled,FALSE),COALESCE(pm.provider,'')
 FROM subscriptions s JOIN plans p ON p.id=s.plan_id
 LEFT JOIN subscription_autopay a ON a.subscription_id=s.id
 LEFT JOIN payment_methods pm ON pm.id=a.payment_method_id
 WHERE s.id=$1 AND s.status='trial' AND s.trial_ends_at IS NOT NULL`, subscriptionID).Scan(&snapshot.Plan, &snapshot.Cycle, &snapshot.EndsAt, &snapshot.AutoDebitEnabled, &snapshot.PaymentMethod); err != nil {
		return err
	}
	snapshot.AutoDebitEnabled = snapshot.AutoDebitEnabled && (snapshot.PaymentMethod == "xendit" || snapshot.PaymentMethod == "midtrans")
	if snapshot.PaymentMethod != "xendit" && snapshot.PaymentMethod != "midtrans" {
		snapshot.PaymentMethod = ""
	}
	snapshot.Amount = amount
	snapshot.Items = items
	raw, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	return queueCheckoutEmail(ctx, tx, "trial_started", subscriptionID, nil, raw)
}

func queueCheckoutEmail(ctx context.Context, tx *sqlx.Tx, event string, subscriptionID int, invoiceID *int, snapshot json.RawMessage) error {
	var owner struct {
		BusinessID string          `db:"business_id"`
		Business   string          `db:"business_name"`
		Customer   string          `db:"customer_name"`
		Recipient  string          `db:"recipient"`
		Settings   json.RawMessage `db:"settings"`
	}
	if err := tx.GetContext(ctx, &owner, `SELECT c.business_id,b.name business_name,c.name customer_name,COALESCE(c.email,'') recipient,COALESCE(b.meta->'invoice','{}'::jsonb) settings FROM subscriptions s JOIN customers c ON c.id=s.customer_id JOIN businesses b ON b.id=c.business_id WHERE s.id=$1`, subscriptionID); err != nil {
		return err
	}
	address, err := mail.ParseAddress(owner.Recipient)
	status := "queued"
	var lastError *string
	if err != nil || address.Address != owner.Recipient {
		if event != "trial_ending" && event != "invoice_due" {
			return nil
		}
		// Keep invalid reminders in history instead of repeatedly selecting them
		// and starving later recipients in the scheduler's bounded batch.
		status = "failed"
		message := "invalid_recipient"
		lastError = &message
	}
	bytes := make([]byte, 32)
	if _, err = rand.Read(bytes); err != nil {
		return err
	}
	if snapshot == nil {
		snapshot = json.RawMessage(`{}`)
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO checkout_email_deliveries(event_type,invoice_id,subscription_id,business_id,business_name,customer_name,recipient,settings,snapshot,public_token,status,last_error) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) ON CONFLICT DO NOTHING`, event, invoiceID, subscriptionID, owner.BusinessID, owner.Business, owner.Customer, owner.Recipient, string(owner.Settings), string(snapshot), base64.RawURLEncoding.EncodeToString(bytes), status, lastError)
	return err
}

func confirmationSubject(kind, business, invoice string) string {
	if kind == "trial_started" {
		return fmt.Sprintf("Trial dimulai - %s", business)
	}
	return fmt.Sprintf("Pembayaran berhasil - %s - %s", invoice, business)
}
