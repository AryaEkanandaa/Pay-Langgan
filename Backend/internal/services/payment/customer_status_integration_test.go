package payment_test

import (
	"errors"
	"testing"

	"pay-langgan/internal/models"
	billingrepo "pay-langgan/internal/repositories/billing"
	customerrepo "pay-langgan/internal/repositories/customer"
	revenuerepo "pay-langgan/internal/repositories/revenue"
	subrepo "pay-langgan/internal/repositories/subscription"
	customer "pay-langgan/internal/services/customer"
	"pay-langgan/internal/utils"
)

func TestCustomerLifecyclePostgres(t *testing.T) {
	db := paymentTestDB(t)
	_, err := db.Exec(`INSERT INTO businesses(id,name) VALUES('other','Other');
 INSERT INTO customers(id,business_id,name,email) VALUES
 (2,'test','Pending','pending@example.com'),(3,'test','Trial','trial@example.com'),
 (4,'test','Cancelled','cancelled@example.com'),(5,'test','Lead','lead@example.com'),(6,'other','Other','other@example.com');
 INSERT INTO subscriptions(id,customer_id,plan_id,status,start_date,trial_ends_at) VALUES
 (2,2,1,'pending',NOW(),NULL),(3,3,1,'trial',NOW(),NOW()+INTERVAL '1 day'),(4,4,1,'cancelled',NOW(),NULL);
 INSERT INTO invoices(id,subscription_id,invoice_number,amount,status) VALUES(2,2,'INV-PENDING',150000,'pending');
 INSERT INTO checkout_sessions(invoice_id,provider,reference_id,public_token_hash,status,expires_at)
 VALUES(2,'xendit','expired-link','test-hash-1','created',(NOW() AT TIME ZONE 'UTC')-INTERVAL '1 hour');
 SELECT setval(pg_get_serial_sequence('customers','id'),6);`)
	if err != nil {
		t.Fatal(err)
	}
	repo := customerrepo.NewCustomerRepository(db)
	svc := customer.NewCustomerService(repo)
	for status, count := range map[string]int{"": 5, "active": 1, "trial": 1, "prospect": 2, "inactive": 1} {
		items, total, err := svc.GetAll("test", 1, 10, "", status)
		if err != nil || total != count || len(items) != count {
			t.Fatalf("filter %q: total=%d items=%d err=%v", status, total, len(items), err)
		}
		for _, item := range items {
			if item.BusinessID != "test" || status != "" && item.Status != status {
				t.Fatal("incorrect scope/status", item.ID)
			}
		}
	}
	items, total, err := svc.GetAll("test", 1, 1, "example.com", "prospect")
	if err != nil || total != 2 || len(items) != 1 {
		t.Fatal("filtered pagination", total, len(items), err)
	}
	if _, _, err := svc.GetAll("test", 1, 10, "", "invalid"); !errors.Is(err, utils.ErrBadRequest) {
		t.Fatal("invalid filter accepted", err)
	}
	invoiceRepo := billingrepo.NewInvoiceRepository(db)
	inv, err := invoiceRepo.FindByIDAndBusinessID(2, "test")
	if err != nil || inv.CheckoutStatus != "expired" || inv.Status != "pending" {
		t.Fatal("expired checkout invoice", inv, err)
	}
	subs, _, err := subrepo.NewSubscriptionRepository(db).FindAllByBusinessID("test", 1, 10, "pending", "")
	if err != nil || len(subs) != 1 || subs[0].CheckoutStatus != "expired" {
		t.Fatal("expired checkout subscription", subs, err)
	}
	// A replacement link clears the expired indication; it does not activate access.
	if _, err = db.Exec(`INSERT INTO checkout_sessions(invoice_id,provider,reference_id,public_token_hash,status,expires_at)
 VALUES(2,'xendit','replacement-link','test-hash-2','created',(NOW() AT TIME ZONE 'UTC')+INTERVAL '1 hour')`); err != nil {
		t.Fatal(err)
	}
	inv, err = invoiceRepo.FindByIDAndBusinessID(2, "test")
	if err != nil || inv.CheckoutStatus != "created" {
		t.Fatal("replacement checkout", inv, err)
	}
	revenue := revenuerepo.NewDashboardRepository(db)
	if mrr, err := revenue.CalculateMRR("test"); err != nil || mrr != 150000 {
		t.Fatal("pending/trial counted in MRR", mrr, err)
	}
	if active, err := revenue.CountActiveCustomers("test"); err != nil || active != 2 {
		t.Fatal("pending customer counted as active", active, err)
	}
	email := " PENDING@EXAMPLE.COM "
	if _, err := svc.Create("test", models.CreateCustomerRequest{Name: "Duplicate", Email: &email}); !errors.Is(err, utils.ErrConflict) {
		t.Fatal("duplicate customer accepted", err)
	}
	if _, err := svc.Update(5, "test", models.UpdateCustomerRequest{Email: &email}); !errors.Is(err, utils.ErrConflict) {
		t.Fatal("duplicate email update accepted", err)
	}
	if _, err := svc.Create("other", models.CreateCustomerRequest{Name: "Independent", Email: &email}); err != nil {
		t.Fatal("cross-tenant email rejected", err)
	}
	if _, err := db.Exec(`UPDATE subscriptions SET status='active' WHERE id=2; UPDATE invoices SET status='paid' WHERE id=2`); err != nil {
		t.Fatal(err)
	}
	items, total, err = svc.GetAll("test", 1, 10, "pending@example.com", "active")
	if err != nil || total != 1 || len(items) != 1 {
		t.Fatal("paid customer did not become active", total, err)
	}
	inv, err = invoiceRepo.FindByIDAndBusinessID(2, "test")
	if err != nil || inv.CheckoutStatus != "" {
		t.Fatal("paid invoice retained checkout warning", inv, err)
	}
}
