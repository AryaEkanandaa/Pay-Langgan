package payment_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	billingrepo "pay-langgan/internal/repositories/billing"
	paymentrepo "pay-langgan/internal/repositories/payment"
	paymentservice "pay-langgan/internal/services/payment"
	"pay-langgan/internal/utils"
	provider "pay-langgan/pkg/payment"
)

func TestMerchantOrderIsScopedAndIdempotent(t *testing.T) {
	db := paymentTestDB(t)
	if _, err := db.Exec(`SELECT setval(pg_get_serial_sequence('customers','id'),(SELECT max(id) FROM customers));
 SELECT setval(pg_get_serial_sequence('subscriptions','id'),(SELECT max(id) FROM subscriptions));
 SELECT setval(pg_get_serial_sequence('invoices','id'),(SELECT max(id) FROM invoices));
 INSERT INTO platform_subscriptions (business_id,status,paid_through)
 VALUES ('test','active',(NOW() AT TIME ZONE 'UTC') + INTERVAL '1 month')`); err != nil {
		t.Fatal(err)
	}
	xendit, err := provider.NewXenditProvider("", "callback-secret", "https://api.xendit.co", true)
	if err != nil {
		t.Fatal(err)
	}
	checkout, err := paymentservice.NewCheckoutService(db, paymentrepo.NewCheckoutRepository(db), xendit,
		"http://localhost:5173", "https://paylanggan.example", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	svc := paymentservice.NewMerchantIntegrationService(db, checkout, billingrepo.NewInvoiceRepository(db))
	ctx := context.Background()
	key, err := svc.CreateKey(ctx, "test", "Demo website")
	if err != nil || !strings.HasPrefix(key.SecretKey, "pl_sk_") {
		t.Fatalf("key: %+v %v", key, err)
	}
	if _, err := svc.Authenticate(ctx, "pl_sk_invalid"); !errors.Is(err, utils.ErrUnauthorized) {
		t.Fatalf("invalid key: %v", err)
	}
	if owner, err := svc.Authenticate(ctx, key.SecretKey); err != nil || owner != "test" {
		t.Fatalf("key owner: %s %v", owner, err)
	}

	req := paymentservice.MerchantOrderRequest{ExternalOrderID: "order-001", PlanID: 1}
	req.Customer.Name = "New Buyer"
	req.Customer.Email = "buyer@example.com"
	req.Customer.Mobile = "+628123456789"
	first, err := svc.CreateOrder(ctx, "test", req)
	if err != nil || first.InvoiceID == 0 || first.CheckoutURL == "" {
		t.Fatalf("first order: %+v %v", first, err)
	}
	second, err := svc.CreateOrder(ctx, "test", req)
	if err != nil || second.InvoiceID != first.InvoiceID || second.CheckoutURL == first.CheckoutURL {
		t.Fatalf("retry: %+v %v", second, err)
	}
	var count int
	if err := db.Get(&count, `SELECT count(*) FROM merchant_orders WHERE business_id='test' AND external_order_id='order-001'`); err != nil || count != 1 {
		t.Fatalf("orders=%d err=%v", count, err)
	}
	if err := db.Get(&count, `SELECT count(*) FROM invoices WHERE id=$1`, first.InvoiceID); err != nil || count != 1 {
		t.Fatalf("invoices=%d err=%v", count, err)
	}
	if _, err := svc.GetOrder(ctx, "other", "order-001"); !errors.Is(err, utils.ErrNotFound) {
		t.Fatalf("cross-tenant order: %v", err)
	}
	if _, err := db.Exec(`UPDATE invoices SET status='paid', paid_at=(NOW() AT TIME ZONE 'UTC') WHERE id=$1`, first.InvoiceID); err != nil {
		t.Fatal(err)
	}
	paidOrder, err := svc.GetOrder(ctx, "test", "order-001")
	if err != nil || paidOrder.InvoiceStatus != "paid" {
		t.Fatalf("paid status: %+v %v", paidOrder, err)
	}
	retryPaid, err := svc.CreateOrder(ctx, "test", req)
	if err != nil || retryPaid.InvoiceID != first.InvoiceID || retryPaid.CheckoutURL != "" {
		t.Fatalf("paid retry: %+v %v", retryPaid, err)
	}

	req.PlanID = 9999
	req.ExternalOrderID = "missing-plan"
	if _, err := svc.CreateOrder(ctx, "test", req); !errors.Is(err, utils.ErrNotFound) {
		t.Fatalf("missing plan: %v", err)
	}
	if _, err := db.Exec(`UPDATE platform_subscriptions SET status='past_due' WHERE business_id='test'`); err != nil {
		t.Fatal(err)
	}
	req.PlanID = 1
	req.ExternalOrderID = "order-after-expiry"
	if _, err := svc.CreateOrder(ctx, "test", req); !errors.Is(err, utils.ErrForbidden) {
		t.Fatalf("inactive platform subscription: %v", err)
	}
	revoked, err := svc.RevokeKey(ctx, "test", key.ID)
	if err != nil || !revoked {
		t.Fatalf("revoke: %v %v", revoked, err)
	}
	if _, err := svc.Authenticate(ctx, key.SecretKey); !errors.Is(err, utils.ErrUnauthorized) {
		t.Fatalf("revoked key: %v", err)
	}
}
