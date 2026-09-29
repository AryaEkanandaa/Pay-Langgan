package platformbilling

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/lib/pq"
	"pay-langgan/internal/config"
	"pay-langgan/internal/database"
	provider "pay-langgan/pkg/payment"
)

type fakeProvider struct {
	sessions map[string]*provider.XenditSessionResponse
	charges  int
}

func (f *fakeProvider) VerifyCallbackToken(token string) bool { return token == "test-token" }
func (f *fakeProvider) CreateHostedSession(_ context.Context, r provider.HostedSessionRequest) (*provider.XenditSessionResponse, error) {
	s := &provider.XenditSessionResponse{PaymentSessionID: "ps-test", ReferenceID: r.ReferenceID, Amount: r.Amount, Currency: "IDR", Mode: "PAYMENT_LINK", Status: "ACTIVE", PaymentLinkURL: "https://test.invalid/pay"}
	f.sessions[s.PaymentSessionID] = s
	return s, nil
}
func (f *fakeProvider) GetHostedSession(_ context.Context, id string) (*provider.XenditSessionResponse, error) {
	return f.sessions[id], nil
}
func (f *fakeProvider) CreateRecurringCharge(_ context.Context, r provider.RecurringChargeRequest) (*provider.RecurringChargeResponse, error) {
	f.charges++
	return &provider.RecurringChargeResponse{PaymentRequestID: "pr-test", ReferenceID: r.ReferenceID, Currency: "IDR", Amount: r.Amount, Status: "SUCCEEDED", PaymentID: "py-renewal"}, nil
}
func (f *fakeProvider) GetRecurringCharge(_ context.Context, _ string) (*provider.RecurringChargeResponse, error) {
	return nil, fmt.Errorf("unexpected poll")
}

func TestPlatformCheckoutAndRenewalPostgres(t *testing.T) {
	if os.Getenv("PLATFORM_BILLING_INTEGRATION") != "1" {
		t.Skip("set PLATFORM_BILLING_INTEGRATION=1 to run isolated PostgreSQL integration test")
	}
	_ = godotenv.Load("../../../.env")
	cfg := config.Load()
	admin, err := database.New(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	schema := fmt.Sprintf("platform_test_%d", time.Now().UnixNano())
	if _, err = admin.Exec("CREATE SCHEMA " + pq.QuoteIdentifier(schema)); err != nil {
		t.Fatal(err)
	}
	defer func() { _, _ = admin.Exec("DROP SCHEMA " + pq.QuoteIdentifier(schema) + " CASCADE") }()
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s search_path=%s", cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode, schema)
	pool, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	db := &database.DB{DB: pool}
	files, err := filepath.Glob("../../../../Migrations/*.sql")
	if err != nil || len(files) == 0 {
		t.Fatal("migrations missing", err)
	}
	for _, file := range files {
		raw, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		up := strings.Split(string(raw), "-- +goose Down")[0]
		if _, err := db.Exec(up); err != nil {
			t.Fatalf("migration %s: %v", filepath.Base(file), err)
		}
	}
	if _, err := db.Exec(`INSERT INTO businesses(id,name) VALUES ('one','One'),('two','Two');
INSERT INTO users(id,business_id,name,email,password,role) VALUES (1,'one','Owner','owner@example.com','unused','admin')`); err != nil {
		t.Fatal(err)
	}
	fake := &fakeProvider{sessions: map[string]*provider.XenditSessionResponse{}}
	svc, err := NewService(db, fake, "https://app.invalid", "https://app.invalid")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if _, err := svc.Start(ctx, "one", 1, false); err == nil {
		t.Fatal("checkout without consent accepted")
	}
	checkout, err := svc.Start(ctx, "one", 1, true)
	if err != nil {
		t.Fatal(err)
	}
	if checkout.Amount != MonthlyPrice || checkout.CheckoutURL == "" {
		t.Fatalf("invalid checkout: %+v", checkout)
	}
	if _, err := svc.Start(ctx, "two", 1, true); err == nil {
		t.Fatal("cross tenant checkout accepted")
	}
	var reference string
	if err := db.Get(&reference, `SELECT reference_id FROM platform_checkout_sessions WHERE invoice_id=$1`, checkout.InvoiceID); err != nil {
		t.Fatal(err)
	}
	event, _ := json.Marshal(map[string]any{"event": "payment_session.completed", "data": map[string]any{
		"payment_session_id": "ps-test", "reference_id": reference, "status": "COMPLETED", "currency": "IDR", "amount": MonthlyPrice, "payment_id": "py-initial", "payment_token_id": "pt-test"}})
	if err := svc.Webhook(ctx, "wrong", event); err == nil {
		t.Fatal("wrong callback token accepted")
	}
	if err := svc.Webhook(ctx, "test-token", event); err != nil {
		t.Fatal(err)
	}
	if err := svc.Webhook(ctx, "test-token", event); err != nil {
		t.Fatal("duplicate webhook:", err)
	}
	status, err := svc.Status(ctx, "one")
	if err != nil || status.Status != "active" || !status.AutoRenew {
		t.Fatalf("invalid paid status: %+v %v", status, err)
	}
	other, err := svc.Status(ctx, "two")
	if err != nil || other.Status != "not_subscribed" {
		t.Fatalf("tenant data leaked: %+v %v", other, err)
	}
	// Advance only the test clock fields; the worker still uses real UTC time.
	if _, err := db.Exec(`UPDATE platform_subscriptions SET billing_anchor=(NOW() AT TIME ZONE 'UTC')-INTERVAL '1 month'+INTERVAL '1 hour',
paid_through=(NOW() AT TIME ZONE 'UTC')-INTERVAL '1 second',
next_billing_date=(NOW() AT TIME ZONE 'UTC')-INTERVAL '1 second',
auto_renew_consent_at=(NOW() AT TIME ZONE 'UTC')-INTERVAL '2 months' WHERE business_id='one'`); err != nil {
		t.Fatal(err)
	}
	first, err := svc.RunOnce(ctx, 10)
	if err != nil {
		t.Fatal(err)
	}
	if first.InvoicesCreated != 1 || first.AutopayStarted != 1 || fake.charges != 1 {
		t.Fatalf("renewal failed: %+v charges=%d", first, fake.charges)
	}
	second, err := svc.RunOnce(ctx, 10)
	if err != nil || second.InvoicesCreated != 0 || second.AutopayStarted != 0 || fake.charges != 1 {
		t.Fatalf("renewal duplicated: %+v err=%v charges=%d", second, err, fake.charges)
	}
	var paid int
	if err := db.Get(&paid, `SELECT count(*) FROM platform_invoices WHERE business_id='one' AND status='paid'`); err != nil {
		t.Fatal(err)
	}
	if paid != 2 {
		t.Fatalf("expected two paid invoices, got %d", paid)
	}
	if err := svc.Cancel(ctx, "one"); err != nil {
		t.Fatal(err)
	}
	status, err = svc.Status(ctx, "one")
	if err != nil || status.Status != "cancelled" || status.AutoRenew {
		t.Fatalf("cancellation failed: %+v %v", status, err)
	}
	if _, err := db.Exec(`UPDATE platform_subscriptions SET paid_through=(NOW() AT TIME ZONE 'UTC')-INTERVAL '1 day' WHERE business_id='one'`); err != nil {
		t.Fatal(err)
	}
	afterCancel, err := svc.RunOnce(ctx, 10)
	if err != nil || afterCancel.InvoicesCreated != 0 || afterCancel.AutopayStarted != 0 || fake.charges != 1 {
		t.Fatalf("debit after cancellation: %+v err=%v charges=%d", afterCancel, err, fake.charges)
	}
}
