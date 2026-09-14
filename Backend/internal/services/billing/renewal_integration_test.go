package billing_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/lib/pq"
	"pay-langgan/internal/config"
	"pay-langgan/internal/database"
	"pay-langgan/internal/models"
	auditrepo "pay-langgan/internal/repositories/audit"
	billingrepo "pay-langgan/internal/repositories/billing"
	"pay-langgan/internal/repositories/catalog"
	couponrepo "pay-langgan/internal/repositories/coupon"
	customerrepo "pay-langgan/internal/repositories/customer"
	subrepo "pay-langgan/internal/repositories/subscription"
	"pay-langgan/internal/services/billing"
	subscriptionsvc "pay-langgan/internal/services/subscription"
)

// Explicit opt-in; only a unique temporary schema is modified. Uses Backend/.env.
func TestRenewalPostgres(t *testing.T) {
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
	schema := fmt.Sprintf("billing_test_%d", time.Now().UnixNano())
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
	exec(t, `INSERT INTO businesses(id,name) VALUES ('test','Test');
  INSERT INTO services(id,business_id,name) VALUES (1,'test','Service');
  INSERT INTO products(id,service_id,name) VALUES (1,1,'Product');
  INSERT INTO plans(id,product_id,name,price,billing_cycle) VALUES (1,1,'Plan',100.01,'monthly');
  INSERT INTO customers(id,business_id,name) VALUES (1,'test','Customer');
  INSERT INTO add_ons(id,product_id,name,price,billing_cycle) VALUES (1,1,'Seats',25.00,'monthly');
  INSERT INTO coupons(id,business_id,code,discount_type,discount_value) VALUES (1,'test','TEN','percentage',10);`)
	newSub := func(t *testing.T, status, anchor, due string, legacy bool) int {
		t.Helper()
		var id int
		var anchorArg, cycleArg interface{} = anchor, "monthly"
		if legacy {
			anchorArg = nil
			cycleArg = nil
		}
		err := db.QueryRow(`INSERT INTO subscriptions(customer_id,plan_id,status,start_date,billing_anchor,billing_cycle,next_billing_date)
   VALUES(1,1,$1,'2026-01-31',$2,$3,$4) RETURNING id`, status, anchorArg, cycleArg, due).Scan(&id)
		if err != nil {
			t.Fatal(err)
		}
		return id
	}
	reset := func(t *testing.T) {
		exec(t, `TRUNCATE subscriptions CASCADE; UPDATE plans SET price=100.01,trial_days=0; UPDATE add_ons SET billing_cycle='monthly'; UPDATE coupons SET expires_at=NULL`)
	}
	cutoff, _ := time.Parse(time.RFC3339, "2026-03-31T00:00:00Z")
	run := func(t *testing.T, limit int) billing.RenewalResult {
		t.Helper()
		result, err := billing.NewRenewalService(db).Run(context.Background(), cutoff, limit)
		if err != nil {
			t.Fatal(err)
		}
		return result
	}
	count := func(t *testing.T, query string) int {
		t.Helper()
		var n int
		if err := db.Get(&n, query); err != nil {
			t.Fatal(err)
		}
		return n
	}

	t.Run("catch up snapshot coupon and replay", func(t *testing.T) {
		reset(t)
		id := newSub(t, "active", "2026-01-31", "2026-02-28", false)
		exec(t, `INSERT INTO subscription_add_ons(subscription_id,add_on_id,quantity) VALUES($1,1,2)`, id)
		exec(t, `INSERT INTO subscription_coupons(subscription_id,coupon_id,applied_at) VALUES($1,1,'2026-01-31')`, id)
		got := run(t, 10)
		if got.InvoicesCreated != 2 || len(got.Failures) != 0 {
			t.Fatalf("%+v", got)
		}
		if run(t, 10).InvoicesCreated != 0 {
			t.Fatal("replay created invoices")
		}
		if count(t, `SELECT count(*) FROM invoices WHERE amount=135.01`) != 2 {
			t.Fatal("incorrect discount rounding")
		}
		if count(t, `SELECT count(*) FROM subscriptions WHERE next_billing_date='2026-04-30'`) != 1 {
			t.Fatal("anchor drift")
		}
		if count(t, `SELECT count(*) FROM invoices i WHERE amount<>(SELECT sum(subtotal) FROM invoice_items WHERE invoice_id=i.id)`) != 0 {
			t.Fatal("items do not reconcile")
		}
		exec(t, `UPDATE plans SET price=999`)
		if count(t, `SELECT count(*) FROM invoice_items WHERE item_type='plan' AND unit_price=100.01`) != 2 {
			t.Fatal("snapshot changed")
		}
		if count(t, `SELECT count(*) FROM audit_logs WHERE action='renew_subscription'`) != 2 {
			t.Fatal("missing audit")
		}
		_, err := db.Exec(`INSERT INTO invoices(subscription_id,invoice_number,amount,status,period_start,period_end) VALUES($1,'duplicate',1,'pending','2026-02-28','2026-03-31')`, id)
		if pg, ok := err.(*pq.Error); !ok || pg.Code != "23505" {
			t.Fatalf("unique constraint missing: %v", err)
		}
	})
	t.Run("concurrent workers", func(t *testing.T) {
		reset(t)
		newSub(t, "active", "2026-01-31", "2026-02-28", false)
		var wg sync.WaitGroup
		results := make(chan error, 2)
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				r, e := billing.NewRenewalService(db).Run(context.Background(), cutoff, 10)
				if e == nil && len(r.Failures) > 0 {
					e = fmt.Errorf("%+v", r.Failures)
				}
				results <- e
			}()
		}
		wg.Wait()
		close(results)
		for err := range results {
			if err != nil {
				t.Fatal(err)
			}
		}
		if count(t, `SELECT count(*) FROM invoices`) != 2 {
			t.Fatal("concurrent duplicate or missing period")
		}
	})
	t.Run("trial paused cancelled legacy and limit", func(t *testing.T) {
		reset(t)
		newSub(t, "trial", "2026-03-31", "2026-03-31", false)
		newSub(t, "paused", "2026-01-31", "2026-02-28", false)
		newSub(t, "cancelled", "2026-01-31", "2026-02-28", false)
		newSub(t, "active", "2026-01-31", "2026-02-28", true)
		got := run(t, 1)
		if got.InvoicesCreated != 1 || got.LegacySkipped != 1 || len(got.Failures) > 0 {
			t.Fatalf("%+v", got)
		}
		if count(t, `SELECT count(*) FROM subscriptions WHERE status='trial'`) != 0 {
			t.Fatal("trial not activated")
		}
		if count(t, `SELECT count(*) FROM invoices`) != 1 {
			t.Fatal("billed excluded subscriptions")
		}
		reset(t)
		newSub(t, "active", "2026-01-31", "2026-02-28", false)
		if run(t, 1).InvoicesCreated != 1 || run(t, 1).InvoicesCreated != 1 {
			t.Fatal("limit did not retain backlog")
		}
	})
	t.Run("failure rolls back and does not starve others", func(t *testing.T) {
		reset(t)
		bad := newSub(t, "active", "2026-01-31", "2026-02-28", false)
		newSub(t, "active", "2026-03-31", "2026-03-31", false)
		// Force an error AFTER invoice insertion to verify the whole transaction.
		exec(t, `CREATE FUNCTION reject_test_item() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN
   IF (SELECT subscription_id FROM invoices WHERE id=NEW.invoice_id)=`+fmt.Sprint(bad)+` THEN RAISE EXCEPTION 'test item failure'; END IF; RETURN NEW; END $$;
   CREATE TRIGGER reject_test_item BEFORE INSERT ON invoice_items FOR EACH ROW EXECUTE FUNCTION reject_test_item()`)
		got := run(t, 10)
		exec(t, `DROP TRIGGER reject_test_item ON invoice_items; DROP FUNCTION reject_test_item()`)
		if got.InvoicesCreated != 1 || len(got.Failures) != 1 {
			t.Fatalf("%+v", got)
		}
		if count(t, fmt.Sprintf(`SELECT count(*) FROM invoices WHERE subscription_id=%d`, bad)) != 0 {
			t.Fatal("partial invoice committed")
		}
		if count(t, fmt.Sprintf(`SELECT count(*) FROM subscriptions WHERE id=%d AND next_billing_date='2026-02-28'`, bad)) != 1 {
			t.Fatal("schedule advanced on failure")
		}
	})
	t.Run("expired coupon and zero invoice", func(t *testing.T) {
		reset(t)
		id := newSub(t, "active", "2026-03-31", "2026-03-31", false)
		exec(t, `INSERT INTO subscription_coupons(subscription_id,coupon_id,applied_at) VALUES($1,1,'2026-01-31')`, id)
		exec(t, `UPDATE coupons SET expires_at='2026-03-31'`)
		if run(t, 10).InvoicesCreated != 1 || count(t, `SELECT count(*) FROM invoices WHERE amount=100.01`) != 1 {
			t.Fatal("expired coupon applied")
		}
		reset(t)
		newSub(t, "active", "2026-03-31", "2026-03-31", false)
		exec(t, `UPDATE plans SET price=0`)
		if run(t, 10).InvoicesCreated != 1 || count(t, `SELECT count(*) FROM invoices WHERE amount=0 AND status='paid' AND paid_at IS NOT NULL`) != 1 {
			t.Fatal("zero invoice not settled")
		}
		if count(t, `SELECT count(*) FROM payments`) != 0 {
			t.Fatal("zero invoice fabricated payment")
		}
	})
	t.Run("initial creation and resume", func(t *testing.T) {
		reset(t)
		// Audit requires a real actor. Match the current role constraints.
		exec(t, `INSERT INTO users(id,business_id,name,email,password,role) VALUES(1,'test','Tester','billing@test.invalid','unused','admin')`)
		service := subscriptionsvc.NewSubscriptionService(db, subrepo.NewSubscriptionRepository(db), subrepo.NewSubscriptionAddOnRepository(db), subrepo.NewSubscriptionCouponRepository(db), auditrepo.NewAuditLogRepository(db), catalog.NewPlanRepository(db), catalog.NewProductRepository(db), catalog.NewServiceRepository(db), catalog.NewAddOnRepository(db), couponrepo.NewCouponRepository(db), customerrepo.NewCustomerRepository(db), billingrepo.NewInvoiceRepository(db))
		_, err := service.Create("test", 1, models.CreateSubscriptionRequest{CustomerID: 1, PlanID: 1})
		if err != nil {
			t.Fatal(err)
		}
		if count(t, `SELECT count(*) FROM invoices WHERE period_start IS NOT NULL AND period_end>period_start`) != 1 || count(t, `SELECT count(*) FROM invoice_items`) != 1 {
			t.Fatal("initial period snapshot missing")
		}
		exec(t, `UPDATE plans SET trial_days=14`)
		if _, err := service.Create("test", 1, models.CreateSubscriptionRequest{CustomerID: 1, PlanID: 1}); err != nil {
			t.Fatal(err)
		}
		if count(t, `SELECT count(*) FROM invoices`) != 1 {
			t.Fatal("trial invoiced early")
		}
		id := newSub(t, "paused", "2026-01-31", "2026-02-28", false)
		if _, err := service.Resume(id, "test", 1); err != nil {
			t.Fatal(err)
		}
		if count(t, fmt.Sprintf(`SELECT count(*) FROM subscriptions WHERE id=%d AND billing_anchor=next_billing_date AND status='active'`, id)) != 1 {
			t.Fatal("resume did not reset anchor")
		}
	})
}
