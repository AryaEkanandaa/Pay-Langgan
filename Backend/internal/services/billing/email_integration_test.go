package billing_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/lib/pq"
	"os"
	"path/filepath"
	"pay-langgan/internal/config"
	"pay-langgan/internal/database"
	billingrepo "pay-langgan/internal/repositories/billing"
	"pay-langgan/internal/services/billing"
	"pay-langgan/internal/utils"
	"strings"
	"sync"
	"testing"
	"time"
)

// Explicit opt-in; only a unique temporary schema is modified. Uses Backend/.env.
func TestInvoiceEmailPostgres(t *testing.T) {
	if os.Getenv("BILLING_INTEGRATION") != "1" {
		t.Skip("set BILLING_INTEGRATION=1 for isolated PostgreSQL tests")
	}
	_ = godotenv.Load("../../../.env")
	cfg := config.Load()
	admin, err := database.New(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)
	if err != nil {
		t.Fatal("connect integration database:", err)
	}
	defer admin.Close()
	schema := fmt.Sprintf("email_test_%d", time.Now().UnixNano())
	if _, err := admin.Exec("CREATE SCHEMA " + pq.QuoteIdentifier(schema)); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, err := admin.Exec("DROP SCHEMA " + pq.QuoteIdentifier(schema) + " CASCADE"); err != nil {
			t.Error(err)
		}
	}()
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s search_path=%s", cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode, schema)
	pool, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatal("connect schema:", err)
	}
	defer pool.Close()
	pool.SetMaxOpenConns(10)
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
	exec := func(t *testing.T, query string, args ...interface{}) {
		t.Helper()
		if _, err := db.Exec(query, args...); err != nil {
			t.Fatal(err)
		}
	}

	exec(t, `INSERT INTO businesses(id,name,meta) VALUES ('test','Kelana','{"invoice":{"auto_send":true,"email_subject":"Invoice {{invoice_number}}"}}'),('other','Other','{}');
 INSERT INTO services(id,business_id,name) VALUES(1,'test','Workspace');
 INSERT INTO products(id,service_id,name) VALUES(1,1,'Desk');
 INSERT INTO plans(id,product_id,name,price,billing_cycle) VALUES(1,1,'Professional',100000,'monthly');
 INSERT INTO customers(id,business_id,name,email) VALUES(1,'test','Customer','customer@kelana.invalid');
 INSERT INTO subscriptions(id,customer_id,plan_id,status,start_date,billing_cycle,billing_anchor) VALUES(1,1,1,'active','2026-10-01','monthly','2026-10-01');`)
	ctx := context.Background()
	repo := billingrepo.NewInvoiceRepository(db)
	svc := billing.NewInvoiceService(db, repo)
	start := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	inv, err := billing.IssueInvoiceContext(ctx, tx, repo, 1, start, start.AddDate(0, 1, 0), "monthly", true)
	if err != nil {
		t.Fatal(err)
	}
	var n int
	if err = tx.Get(&n, `SELECT count(*) FROM invoice_email_deliveries`); err != nil || n != 1 {
		t.Fatalf("auto queue: %d %v", n, err)
	}
	tx.Rollback()
	db.Get(&n, `SELECT count(*) FROM invoice_email_deliveries`)
	if n != 0 {
		t.Fatal("job survived invoice rollback")
	}
	tx, err = db.BeginTxx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	inv, err = billing.IssueInvoiceContext(ctx, tx, repo, 1, start, start.AddDate(0, 1, 0), "monthly", true)
	if err != nil {
		t.Fatal(err)
	}
	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}
	if _, err = svc.QueueEmail(ctx, inv.ID, "other"); !errors.Is(err, utils.ErrNotFound) {
		t.Fatal("tenant access allowed", err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, e := svc.QueueEmail(ctx, inv.ID, "test"); e != nil {
				t.Error(e)
			}
		}()
	}
	wg.Wait()
	db.Get(&n, `SELECT count(*) FROM invoice_email_deliveries`)
	if n != 1 {
		t.Fatal("duplicate jobs")
	}
	cfg.EmailDeliveryMode = "preview"
	cfg.EmailPreviewDir = t.TempDir()
	cfg.APIPublicURL = "https://backend.kelana.invalid"
	if err = svc.RunEmailBatch(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	state, err := svc.EmailStatus(ctx, inv.ID, "test")
	if err != nil || state.Status != "previewed" || state.SentAt != nil {
		t.Fatal("preview claimed sent", state, err)
	}
	content, err := os.ReadFile(filepath.Join(cfg.EmailPreviewDir, fmt.Sprintf("invoice-%d.html", inv.ID)))
	if err != nil || !strings.Contains(string(content), "Rp 100.000") {
		t.Fatal("preview file invalid", err)
	}
	// A later settings edit must not change the appearance of a queued email.
	var previewName, previewCustomer string
	var snapshot, previewSettings json.RawMessage
	if err = db.Get(&snapshot, `SELECT settings FROM invoice_email_deliveries WHERE invoice_id=$1`, inv.ID); err != nil {
		t.Fatal(err)
	}
	exec(t, `UPDATE businesses SET meta=jsonb_set(meta,'{invoice,footer_note}','"edited after queue"') WHERE id='test'`)
	if err = svc.EmailPreviewSettings(ctx, inv.ID, "test", &previewName, &previewCustomer, &previewSettings); err != nil {
		t.Fatal(err)
	}
	if string(snapshot) != string(previewSettings) || previewName == "" || previewCustomer == "" {
		t.Fatal("preview did not preserve delivery snapshot")
	}
	var token string
	db.Get(&token, `SELECT public_token FROM invoice_email_deliveries WHERE invoice_id=$1`, inv.ID)
	if _, _, _, err = svc.PublicEmailInvoice(ctx, token); err != nil {
		t.Fatal(err)
	}
	exec(t, `UPDATE invoice_email_deliveries SET expires_at=NOW()-INTERVAL '1 second'`)
	if _, _, _, err = svc.PublicEmailInvoice(ctx, token); !errors.Is(err, utils.ErrNotFound) {
		t.Fatal("expired link allowed")
	}
	exec(t, `UPDATE businesses SET meta='{"invoice":{"auto_send":false}}' WHERE id='test'`)
	tx, _ = db.BeginTxx(ctx, nil)
	_, err = billing.IssueInvoiceContext(ctx, tx, repo, 1, start.AddDate(0, 1, 0), start.AddDate(0, 2, 0), "monthly", false)
	if err != nil {
		t.Fatal(err)
	}
	tx.Commit()
	db.Get(&n, `SELECT count(*) FROM invoice_email_deliveries`)
	if n != 1 {
		t.Fatal("auto off ignored")
	}
	// An explicitly retried preview uses the current business settings and recipient.
	exec(t, `UPDATE customers SET email='updated@kelana.invalid' WHERE id=1`)
	state, err = svc.QueueEmail(ctx, inv.ID, "test")
	if err != nil || state.Recipient != "updated@kelana.invalid" || state.Status != "queued" {
		t.Fatal("retry stale", state, err)
	}
	// Failing preview storage retries, and the final lease recovers to failed.
	cfg.EmailPreviewDir = filepath.Join(cfg.EmailPreviewDir, "block")
	if err = os.WriteFile(cfg.EmailPreviewDir, []byte("block"), 0600); err != nil {
		t.Fatal(err)
	}
	if err = svc.RunEmailBatch(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	state, _ = svc.EmailStatus(ctx, inv.ID, "test")
	if state.Status != "queued" || state.LastError == nil {
		t.Fatal("failure retry missing")
	}
	exec(t, `UPDATE invoice_email_deliveries SET status='sending',attempts=5,next_attempt_at=NOW()-INTERVAL '1 second'`)
	if err = svc.RunEmailBatch(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	state, _ = svc.EmailStatus(ctx, inv.ID, "test")
	if state.Status != "failed" {
		t.Fatal("final lease stuck")
	}
}
