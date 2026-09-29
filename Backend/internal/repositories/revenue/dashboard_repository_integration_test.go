package revenue

import (
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
)

// Run with REVENUE_INTEGRATION=1 against a local PostgreSQL instance.
func TestDashboardAggregatesPostgres(t *testing.T) {
	if os.Getenv("REVENUE_INTEGRATION") != "1" {
		t.Skip("set REVENUE_INTEGRATION=1 to run PostgreSQL aggregate test")
	}
	_ = godotenv.Load("../../../.env")
	cfg := config.Load()
	admin, err := database.New(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	schema := fmt.Sprintf("revenue_test_%d", time.Now().UnixNano())
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
		raw, readErr := os.ReadFile(file)
		if readErr != nil {
			t.Fatal(readErr)
		}
		if _, err = db.Exec(strings.Split(string(raw), "-- +goose Down")[0]); err != nil {
			t.Fatalf("migration %s: %v", filepath.Base(file), err)
		}
	}
	_, err = db.Exec(`
 INSERT INTO businesses(id,name) VALUES ('one','One'),('two','Two');
 INSERT INTO services(id,business_id,name) VALUES (1,'one','Service One'),(2,'two','Service Two');
 INSERT INTO products(id,service_id,name) VALUES (1,1,'Product One'),(2,2,'Product Two');
 INSERT INTO plans(id,product_id,name,price,billing_cycle) VALUES (1,1,'Monthly',100000,'monthly'),(2,1,'Trial',50000,'monthly'),(3,2,'Other',900000,'monthly');
 INSERT INTO customers(id,business_id,name) VALUES (1,'one','Customer One'),(2,'one','Customer Trial'),(3,'two','Customer Other');
 INSERT INTO subscriptions(id,customer_id,plan_id,status,start_date,created_at) VALUES
  (1,1,1,'active',NOW(),NOW() AT TIME ZONE 'UTC'),
  (2,2,2,'trial',NOW(),NOW() AT TIME ZONE 'UTC'),
  (3,3,3,'active',NOW(),NOW() AT TIME ZONE 'UTC');
 INSERT INTO invoices(id,subscription_id,invoice_number,amount,status,due_date) VALUES
  (1,1,'ONE-PAID',100000,'paid',NOW() AT TIME ZONE 'UTC'),
  (2,1,'ONE-DUE',100000,'pending',(NOW() AT TIME ZONE 'UTC')-interval '1 day'),
  (3,3,'TWO-PAID',900000,'paid',NOW() AT TIME ZONE 'UTC');
 INSERT INTO payments(id,invoice_id,provider,amount,status,paid_at) VALUES
  (1,1,'xendit',100000,'success',NOW() AT TIME ZONE 'UTC'),
  (2,3,'xendit',900000,'success',NOW() AT TIME ZONE 'UTC');
 INSERT INTO platform_subscriptions(business_id,status) VALUES ('one','active'),('two','past_due');
 INSERT INTO platform_invoices(business_id,period_start,period_end,amount,status,due_date,paid_at) VALUES
  ('one',(NOW() AT TIME ZONE 'UTC')-interval '1 month',NOW() AT TIME ZONE 'UTC',300000,'paid',NOW() AT TIME ZONE 'UTC',NOW() AT TIME ZONE 'UTC'),
  ('two',NOW() AT TIME ZONE 'UTC',(NOW() AT TIME ZONE 'UTC')+interval '1 month',300000,'pending',NOW() AT TIME ZONE 'UTC',NULL);
 `)
	if err != nil {
		t.Fatal(err)
	}
	repo := NewDashboardRepository(db)
	mrr, err := repo.CalculateMRR("one")
	if err != nil || mrr != 100000 {
		t.Fatalf("MRR = %v, err = %v; trial should be excluded", mrr, err)
	}
	financials, err := repo.Financials("one")
	if err != nil || financials.RevenueThisMonth != 100000 || financials.OutstandingAmount != 100000 || financials.OverdueAmount != 100000 || financials.PaymentsSucceeded != 1 || financials.NewSubscriptions != 2 {
		t.Fatalf("tenant financials = %+v, err = %v", financials, err)
	}
	plans, err := repo.RevenueByPlan("one")
	if err != nil || len(plans) != 1 || plans[0].PlanID != 1 || plans[0].Amount != 100000 {
		t.Fatalf("plan revenue = %+v, err = %v", plans, err)
	}
	platform, err := repo.PlatformSummary()
	if err != nil || platform.ActiveBusinesses != 1 || platform.PastDueBusinesses != 1 || platform.RevenueThisMonth != 300000 || platform.OutstandingAmount != 300000 {
		t.Fatalf("platform summary = %+v, err = %v", platform, err)
	}
}
