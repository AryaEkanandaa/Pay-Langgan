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
	req.ExternalOrderID = "order-002"
	req.Customer.Email = "BUYER@example.com"
	anotherOrder, err := svc.CreateOrder(ctx, "test", req)
	if err != nil || anotherOrder.InvoiceID == first.InvoiceID {
		t.Fatalf("second purchase: %+v %v", anotherOrder, err)
	}
	if err := db.Get(&count, `SELECT count(*) FROM customers WHERE business_id='test' AND lower(email)='buyer@example.com'`); err != nil || count != 1 {
		t.Fatalf("customer reuse count=%d err=%v", count, err)
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

func TestMerchantEmbedSiteAcceptsNewPlansAndRestrictsOriginAndRevocation(t *testing.T) {
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
	if _, err := svc.CreateEmbedSite(ctx, "test", paymentservice.CreateEmbedSiteRequest{
		Name: "Wrong website", Origin: "https://merchant.example/path",
	}); !errors.Is(err, utils.ErrBadRequest) {
		t.Fatalf("path origin: %v", err)
	}
	site, err := svc.CreateEmbedSite(ctx, "test", paymentservice.CreateEmbedSiteRequest{})
	if err != nil || !strings.HasPrefix(site.PublicID, "pl_site_") || site.Name != "Integrasi bisnis" || site.Origin != "" {
		t.Fatalf("site: %+v %v", site, err)
	}
	sites, err := svc.ListEmbedSites(ctx, "test")
	if err != nil || len(sites) != 1 || sites[0].PublicID != site.PublicID {
		t.Fatalf("sites: %+v %v", sites, err)
	}
	var firstPublicID string
	if err := db.Get(&firstPublicID, `SELECT public_id FROM plans WHERE id=1`); err != nil || !strings.HasPrefix(firstPublicID, "pl_plan_") {
		t.Fatalf("first plan public ID: %q %v", firstPublicID, err)
	}
	request := paymentservice.EmbedOrderRequest{PlanPublicID: firstPublicID}
	request.Customer.Name = "Dira Pratama"
	request.Customer.Email = "dira@merchant.example"
	request.Customer.Mobile = "+628123456789"
	request.PlanPublicID = "pl_plan_ffffffffffffffffffffffff"
	if _, err := svc.CreateEmbedOrder(ctx, site.PublicID, "https://any-origin.example", request); !errors.Is(err, utils.ErrNotFound) {
		t.Fatalf("missing plan: %v", err)
	}
	request.PlanPublicID = firstPublicID
	order, err := svc.CreateEmbedOrder(ctx, site.PublicID, "https://any-origin.example", request)
	if err != nil || order.InvoiceID == 0 || order.CheckoutURL == "" {
		t.Fatalf("order: %+v %v", order, err)
	}
	if _, err := db.Exec(`INSERT INTO plans(id,product_id,name,price,billing_cycle) VALUES (2,1,'Team',200000,'monthly')`); err != nil {
		t.Fatal(err)
	}
	var newPublicID string
	if err := db.Get(&newPublicID, `SELECT public_id FROM plans WHERE id=2`); err != nil || newPublicID == firstPublicID {
		t.Fatalf("new plan public ID: %q %v", newPublicID, err)
	}
	if _, err := db.Exec(`UPDATE plans SET name='Team Plus' WHERE id=2`); err != nil {
		t.Fatal(err)
	}
	request.PlanPublicID = newPublicID
	order, err = svc.CreateEmbedOrder(ctx, site.PublicID, "https://any-origin.example", request)
	if err != nil || order.InvoiceID == 0 {
		t.Fatalf("new plan should work without updating website: %+v %v", order, err)
	}
	request.PlanPublicID = ""
	request.PlanID = 1
	legacyOrder, err := svc.CreateEmbedOrder(ctx, site.PublicID, "https://any-origin.example", request)
	if err != nil || legacyOrder.InvoiceID == 0 {
		t.Fatalf("old numeric plan button should still work: %+v %v", legacyOrder, err)
	}
	if revoked, err := svc.RevokeEmbedSite(ctx, "test", site.ID); err != nil || !revoked {
		t.Fatalf("revoke: %v %v", revoked, err)
	}
	if _, err := svc.CreateEmbedOrder(ctx, site.PublicID, "https://any-origin.example", request); !errors.Is(err, utils.ErrNotFound) {
		t.Fatalf("revoked website: %v", err)
	}
}
