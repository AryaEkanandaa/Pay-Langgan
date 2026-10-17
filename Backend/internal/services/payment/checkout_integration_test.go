package payment_test

import (
	"context"
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
	"pay-langgan/internal/models"
	paymentrepo "pay-langgan/internal/repositories/payment"
	paymentservice "pay-langgan/internal/services/payment"
	provider "pay-langgan/pkg/payment"
)

func paymentTestDB(t *testing.T) *database.DB {
	t.Helper()
	if os.Getenv("PAYMENT_INTEGRATION") != "1" {
		t.Skip("set PAYMENT_INTEGRATION=1 for isolated PostgreSQL tests")
	}
	_ = godotenv.Load("../../../.env")
	cfg := config.Load()
	admin, err := database.New(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)
	if err != nil {
		t.Fatal("connect integration database:", err)
	}
	t.Cleanup(func() { admin.Close() })
	schema := fmt.Sprintf("payment_test_%d", time.Now().UnixNano())
	if _, err := admin.Exec("CREATE SCHEMA " + pq.QuoteIdentifier(schema)); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = admin.Exec("DROP SCHEMA " + pq.QuoteIdentifier(schema) + " CASCADE")
	})
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s search_path=%s", cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode, schema)
	pool, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { pool.Close() })
	db := &database.DB{DB: pool}
	files, err := filepath.Glob("../../../../Migrations/*.sql")
	if err != nil || len(files) == 0 {
		t.Fatal("migrations missing", err)
	}
	for _, file := range files {
		raw, readErr := os.ReadFile(file)
		if readErr != nil {
			t.Fatal(readErr)
		}
		up := strings.Split(string(raw), "-- +goose Down")[0]
		if _, err := db.Exec(up); err != nil {
			t.Fatalf("migration %s: %v", filepath.Base(file), err)
		}
	}
	_, err = db.Exec(`INSERT INTO businesses(id,name) VALUES ('test','Studio Test');
 INSERT INTO services(id,business_id,name) VALUES (1,'test','Service');
 INSERT INTO products(id,service_id,name) VALUES (1,1,'Product');
 INSERT INTO plans(id,product_id,name,price,billing_cycle) VALUES (1,1,'Pro',150000,'monthly');
 INSERT INTO customers(id,business_id,name,email,contact) VALUES (1,'test','Alya','alya@example.com','+628123456789');
 INSERT INTO subscriptions(id,customer_id,plan_id,status,start_date) VALUES (1,1,1,'active',NOW());
 INSERT INTO invoices(id,subscription_id,invoice_number,amount,status) VALUES (1,1,'INV-TEST',150000,'pending')`)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func TestCheckoutPostgres(t *testing.T) {
	db := paymentTestDB(t)
	if _, err := db.Exec(`UPDATE subscriptions SET status='pending' WHERE id=1`); err != nil {
		t.Fatal(err)
	}
	xendit, err := provider.NewXenditProvider("", "callback-secret", "https://api.xendit.co", true)
	if err != nil {
		t.Fatal(err)
	}
	service, err := paymentservice.NewCheckoutService(db, paymentrepo.NewCheckoutRepository(db), xendit,
		"http://localhost:5173", "https://paylanggan.example", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	link, err := service.CreateLink(ctx, 1, "test")
	if err != nil || link.Token == "" {
		t.Fatalf("link=%+v err=%v", link, err)
	}
	replacement, err := service.CreateLink(ctx, 1, "test")
	if err != nil || replacement.Token == link.Token {
		t.Fatalf("replacement=%+v err=%v", replacement, err)
	}
	oldCheckout, err := service.Get(ctx, link.Token)
	if err != nil || oldCheckout.CheckoutStatus != "expired" {
		t.Fatalf("old checkout=%+v err=%v", oldCheckout, err)
	}
	link = replacement
	checkout, err := service.Get(ctx, link.Token)
	if err != nil || checkout.InvoiceNumber != "INV-TEST" {
		t.Fatalf("checkout=%+v err=%v", checkout, err)
	}
	session, err := service.Start(ctx, link.Token, models.StartCheckoutRequest{
		Name: "Alya", Email: "alya@example.com", Mobile: "+628123456789",
	})
	if err != nil || session.ComponentsSDKKey == "" || session.Mode != "COMPONENTS" {
		t.Fatalf("session=%+v err=%v", session, err)
	}
	resumed, err := service.Start(ctx, link.Token, models.StartCheckoutRequest{})
	if err != nil || resumed.PaymentSessionID != session.PaymentSessionID || resumed.ComponentsSDKKey != session.ComponentsSDKKey {
		t.Fatal("failed to resume components checkout", err)
	}
	var payments int
	if err := db.Get(&payments, `SELECT count(*) FROM payments WHERE invoice_id=1`); err != nil || payments != 1 {
		t.Fatal("resume created another payment", err)
	}
	if _, err := service.CreateLink(ctx, 1, "test"); err == nil {
		t.Fatal("second link was created while a Xendit session was active")
	}
	var providerSessionID, referenceID string
	if err := db.QueryRow(`SELECT provider_session_id, reference_id FROM checkout_sessions WHERE invoice_id=1 AND status='active'`).Scan(&providerSessionID, &referenceID); err != nil {
		t.Fatal(err)
	}
	payload := fmt.Sprintf(`{"event":"payment_session.completed","data":{"payment_session_id":%q,"reference_id":%q,"status":"COMPLETED","currency":"IDR","amount":150000,"payment_token_id":"pt-test","payment_request_id":"pr-test","payment_id":"py-test"}}`, providerSessionID, referenceID)
	duplicate, err := service.ProcessWebhook(ctx, "callback-secret", []byte(payload))
	if err != nil || duplicate {
		t.Fatalf("first webhook duplicate=%v err=%v", duplicate, err)
	}
	duplicate, err = service.ProcessWebhook(ctx, "callback-secret", []byte(payload))
	if err != nil || !duplicate {
		t.Fatalf("replayed webhook duplicate=%v err=%v", duplicate, err)
	}
	var invoiceStatus, paymentStatus, paymentID string
	if err := db.QueryRow(`SELECT i.status, p.status, p.provider_transaction_id FROM invoices i JOIN payments p ON p.invoice_id=i.id WHERE i.id=1`).Scan(&invoiceStatus, &paymentStatus, &paymentID); err != nil {
		t.Fatal(err)
	}
	if invoiceStatus != "paid" || paymentStatus != "success" || paymentID != "py-test" {
		t.Fatalf("invoice=%s payment=%s id=%s", invoiceStatus, paymentStatus, paymentID)
	}
	var methods, events int
	_ = db.Get(&methods, `SELECT count(*) FROM payment_methods WHERE token='pt-test'`)
	_ = db.Get(&events, `SELECT count(*) FROM webhook_events WHERE processed=TRUE`)
	if methods != 1 || events != 1 {
		t.Fatalf("methods=%d events=%d", methods, events)
	}
	var confirmations int
	var subscriptionStatus string
	if err := db.Get(&subscriptionStatus, `SELECT status FROM subscriptions WHERE id=1`); err != nil || subscriptionStatus != "active" {
		t.Fatalf("Xendit subscription=%s %v", subscriptionStatus, err)
	}
	if err := db.Get(&confirmations, `SELECT count(*) FROM checkout_email_deliveries WHERE invoice_id=1 AND event_type='payment_succeeded'`); err != nil || confirmations != 1 {
		t.Fatalf("Xendit confirmations=%d %v", confirmations, err)
	}
}
