package revenue

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/lib/pq"
	"os"
	"pay-langgan/internal/config"
	"pay-langgan/internal/database"
	"testing"
	"time"
)

func TestDashboardRangePostgres(t *testing.T) {
	if os.Getenv("DASHBOARD_INTEGRATION") != "1" {
		t.Skip("set DASHBOARD_INTEGRATION=1; isolated temporary schema only")
	}
	_ = godotenv.Load("../../../.env")
	cfg := config.Load()
	admin, err := database.New(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	schema := fmt.Sprintf("dashboard_test_%d", time.Now().UnixNano())
	if _, err = admin.Exec("CREATE SCHEMA " + pq.QuoteIdentifier(schema)); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, err := admin.Exec("DROP SCHEMA " + pq.QuoteIdentifier(schema) + " CASCADE"); err != nil {
			t.Error(err)
		}
	}()
	pool, err := sqlx.Connect("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s search_path=%s", cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode, schema))
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	db := &database.DB{DB: pool}
	_, err = db.Exec(`CREATE TABLE businesses(id text,meta jsonb);
 CREATE TABLE customers(id int,business_id text);
 CREATE TABLE subscriptions(id int,customer_id int,plan_id int,created_at timestamp,end_date timestamp,status text);
 CREATE TABLE plans(id int,name text);
 CREATE TABLE invoices(id int,subscription_id int);
 CREATE TABLE payments(id int,invoice_id int,paid_at timestamp,created_at timestamp,status text,amount numeric);
 CREATE TABLE autopay_charges(payment_id int,status text,updated_at timestamp);
 INSERT INTO businesses VALUES ('tenant','{"timezone":"Asia/Makassar"}');
 INSERT INTO customers VALUES(1,'tenant'),(2,'other');
 INSERT INTO plans VALUES(1,'Plan');
 INSERT INTO subscriptions VALUES(1,1,1,'2026-10-05 16:00','2026-10-06 15:59','cancelled'),(2,2,1,'2026-10-05 16:00',NULL,'active');
 INSERT INTO invoices VALUES(1,1),(2,2);
 INSERT INTO payments VALUES(1,1,'2026-10-05 15:59','2026-10-05 15:59','success',500),
 (2,1,'2026-10-05 16:00','2026-10-05 16:00','success',100000),
 (3,1,'2026-10-06 16:00','2026-10-06 16:00','success',999),
 (4,2,'2026-10-05 16:00','2026-10-05 16:00','success',999999),
 (5,1,NULL,'2026-10-06 15:59','failed',100000);
 INSERT INTO autopay_charges VALUES(2,'succeeded','2026-10-05 16:00'),(5,'failed','2026-10-06 15:59');`)
	if err != nil {
		t.Fatal(err)
	}
	repo := NewDashboardRepository(db)
	start := time.Date(2026, 10, 5, 16, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)
	for _, grain := range []string{"day", "week", "month"} {
		rows, err := repo.RangeSeries("tenant", start, end, "Asia/Makassar", grain, false)
		if err != nil {
			t.Fatal(err)
		}
		if len(rows) != 1 || rows[0].Revenue != 100000 || rows[0].PaymentsSucceeded != 1 || rows[0].PaymentsFailed != 1 || rows[0].AutopaySucceeded != 1 || rows[0].AutopayFailed != 1 || rows[0].NewSubscriptions != 1 || rows[0].CancelledSubscriptions != 1 {
			t.Fatalf("incorrect %s stats: %+v", grain, rows)
		}
	}
	rows, err := repo.RangeSeries("tenant", start, end, "Asia/Makassar", "day", true)
	if err != nil || len(rows) != 1 || rows[0].Revenue != 0 || rows[0].PaymentsSucceeded != 0 || rows[0].NewSubscriptions != 1 {
		t.Fatalf("sales financial leak: %+v %v", rows, err)
	}
	plans, err := repo.RangePlanRevenue("tenant", start, end)
	if err != nil || len(plans) != 1 || plans[0].Amount != 100000 {
		t.Fatalf("plan totals: %+v %v", plans, err)
	}
}
