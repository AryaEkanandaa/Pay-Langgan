package payment_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"pay-langgan/internal/config"
	"pay-langgan/internal/models"
	catalogrepo "pay-langgan/internal/repositories/catalog"
	catalogsvc "pay-langgan/internal/services/catalog"
	"strings"
	"sync"
	"testing"
	"time"

	billingrepo "pay-langgan/internal/repositories/billing"
	paymentrepo "pay-langgan/internal/repositories/payment"
	billing "pay-langgan/internal/services/billing"
	payment "pay-langgan/internal/services/payment"
	"pay-langgan/internal/utils"
	provider "pay-langgan/pkg/payment"
)

func TestMerchantCheckoutOptionsAndTrial(t *testing.T) {
	db := paymentTestDB(t)
	_, err := db.Exec(`SELECT setval(pg_get_serial_sequence('customers','id'),(SELECT max(id) FROM customers));
 SELECT setval(pg_get_serial_sequence('subscriptions','id'),(SELECT max(id) FROM subscriptions));
 SELECT setval(pg_get_serial_sequence('invoices','id'),(SELECT max(id) FROM invoices));
 INSERT INTO platform_subscriptions(business_id,status,paid_through) VALUES('test','active',NOW()+INTERVAL '1 month');
 INSERT INTO add_ons(id,product_id,name,price,billing_cycle) VALUES(1,1,'Extra seats',25000,'monthly'),(2,1,'Yearly seats',25000,'yearly');
 INSERT INTO businesses(id,name) VALUES('another','Other tenant');
 INSERT INTO services(id,business_id,name) VALUES(2,'another','Other service');
 INSERT INTO products(id,service_id,name) VALUES(2,2,'Other product');
 INSERT INTO add_ons(id,product_id,name,price,billing_cycle) VALUES(3,2,'Private addon',10000,'monthly');
 INSERT INTO coupons(id,business_id,code,discount_type,discount_value,max_usage) VALUES(1,'test','SAVE10','percentage',10,10),(2,'test','ONCE','fixed',10000,1),(3,'test','FREE','percentage',100,10),(4,'another','PRIVATE','fixed',5000,5);
 INSERT INTO plans(id,product_id,name,price,billing_cycle,trial_days) VALUES(2,1,'Trial Pro',100000,'monthly',7);`)
	if err != nil {
		t.Fatal(err)
	}
	baseGateway, _ := provider.NewXenditProvider("", "test-secret", "https://api.xendit.co", true)
	gateway := &trialSessionProvider{SessionProvider: baseGateway}
	checkout, err := payment.NewCheckoutService(db, paymentrepo.NewCheckoutRepository(db), gateway, "http://localhost:5173", "https://paylanggan.example", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	svc := payment.NewMerchantIntegrationService(db, checkout, billingrepo.NewInvoiceRepository(db))
	ctx := context.Background()
	site, err := svc.CreateEmbedSite(ctx, "test", payment.CreateEmbedSiteRequest{})
	if err != nil {
		t.Fatal(err)
	}
	var publicPlan string
	db.Get(&publicPlan, `SELECT public_id FROM plans WHERE id=1`)
	options := []payment.CheckoutAddOn{{AddOnID: 1, Quantity: 2}}
	// Catalog visibility persists and is enforced for both public quotes and orders.
	catalog := catalogsvc.NewAddOnService(catalogrepo.NewAddOnRepository(db), catalogrepo.NewProductRepository(db))
	hidden := false
	if _, err := db.Exec(`SELECT setval(pg_get_serial_sequence('add_ons','id'),(SELECT max(id) FROM add_ons))`); err != nil {
		t.Fatal(err)
	}
	created, err := catalog.Create("test", models.CreateAddOnRequest{ProductID: 1, Name: "Admin only", Price: 10000, BillingCycle: "monthly", VisibleInCheckout: &hidden})
	if err != nil || created.VisibleInCheckout {
		t.Fatalf("create hidden addon: %+v %v", created, err)
	}
	if _, err := catalog.Update(created.ID, "test", models.UpdateAddOnRequest{Name: "Admin only updated"}); err != nil {
		t.Fatal(err)
	}
	stored, err := catalog.GetByID(created.ID, "test")
	if err != nil || stored.VisibleInCheckout {
		t.Fatalf("omitted visibility must preserve hidden state: %+v %v", stored, err)
	}
	addon, err := catalog.Update(1, "test", models.UpdateAddOnRequest{VisibleInCheckout: &hidden})
	if err != nil || addon.VisibleInCheckout {
		t.Fatalf("hide addon: %+v %v", addon, err)
	}
	listed, err := svc.QuoteEmbedPlan(ctx, site.PublicID, publicPlan, nil, "")
	if err != nil || len(listed.AddOns) != 0 {
		t.Fatalf("hidden addon listed: %+v %v", listed, err)
	}
	if _, err := svc.QuoteEmbedPlan(ctx, site.PublicID, publicPlan, options, ""); !errors.Is(err, utils.ErrBadRequest) {
		t.Fatalf("hidden addon accepted by quote: %v", err)
	}
	hiddenRequest := payment.MerchantOrderRequest{ExternalOrderID: "hidden-addon", PlanID: 1, AddOns: options}
	hiddenRequest.Customer.Name = "Buyer"
	hiddenRequest.Customer.Email = "hidden@example.invalid"
	hiddenRequest.Customer.Mobile = "+628123456789"
	if _, err := svc.CreateOrder(ctx, "test", hiddenRequest); !errors.Is(err, utils.ErrBadRequest) {
		t.Fatalf("hidden addon accepted by order: %v", err)
	}
	visible := true
	if _, err := catalog.Update(1, "test", models.UpdateAddOnRequest{VisibleInCheckout: &visible}); err != nil {
		t.Fatal(err)
	}
	quote, err := svc.QuoteEmbedPlan(ctx, site.PublicID, publicPlan, options, "SAVE10")
	if err != nil || quote.Amount != 180000 || quote.DueNow != 180000 || len(quote.Items) != 3 || len(quote.AddOns) != 1 {
		t.Fatalf("quote=%+v err=%v", quote, err)
	}
	makeRequest := func(ref, email string) payment.MerchantOrderRequest {
		r := payment.MerchantOrderRequest{ExternalOrderID: ref, PlanID: 1, AddOns: options, CouponCode: "SAVE10"}
		r.Customer.Name = "Buyer"
		r.Customer.Email = email
		r.Customer.Mobile = "+628123456789"
		return r
	}
	req := makeRequest("options-order", "options@example.invalid")
	order, err := svc.CreateOrder(ctx, "test", req)
	if err != nil {
		t.Fatal(err)
	}
	if order.SubscriptionStatus != "pending" {
		t.Fatalf("unpaid order active: %+v", order)
	}
	var amount float64
	db.Get(&amount, `SELECT amount FROM invoices WHERE id=$1`, order.InvoiceID)
	if amount != quote.Amount {
		t.Fatalf("invoice=%v quote=%v", amount, quote.Amount)
	}
	retry, err := svc.CreateOrder(ctx, "test", req)
	if err != nil || retry.InvoiceID != order.InvoiceID {
		t.Fatalf("retry %+v %v", retry, err)
	}
	req.AddOns = []payment.CheckoutAddOn{{AddOnID: 1, Quantity: 3}}
	if _, err = svc.CreateOrder(ctx, "test", req); !errors.Is(err, utils.ErrConflict) {
		t.Fatalf("changed request must conflict: %v", err)
	}
	for _, bad := range [][]payment.CheckoutAddOn{{{AddOnID: 2, Quantity: 1}}, {{AddOnID: 3, Quantity: 1}}, {{AddOnID: 1, Quantity: 0}}, {{AddOnID: 1, Quantity: 1}, {AddOnID: 1, Quantity: 1}}} {
		if _, err = svc.QuoteEmbedPlan(ctx, site.PublicID, publicPlan, bad, ""); !errors.Is(err, utils.ErrBadRequest) {
			t.Fatalf("invalid addons accepted: %+v %v", bad, err)
		}
	}
	for _, code := range []string{"PRIVATE", "MISSING"} {
		if _, err = svc.QuoteEmbedPlan(ctx, site.PublicID, publicPlan, nil, code); !errors.Is(err, utils.ErrBadRequest) {
			t.Fatalf("invalid coupon %s: %v", code, err)
		}
	}
	// Concurrent orders must not spend the last coupon slot twice.
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for i, ref := range []string{"quota-a", "quota-b"} {
		wg.Add(1)
		go func(ref string, i int) {
			defer wg.Done()
			r := makeRequest(ref, ref+"@example.invalid")
			r.CouponCode = "ONCE"
			_, e := svc.CreateOrder(ctx, "test", r)
			results <- e
		}(ref, i)
	}
	wg.Wait()
	close(results)
	success := 0
	for e := range results {
		if e == nil {
			success++
		} else if !errors.Is(e, utils.ErrBadRequest) {
			t.Fatal(e)
		}
	}
	if success != 1 {
		t.Fatalf("coupon quota successes=%d", success)
	}
	free := makeRequest("free-order", "free@example.invalid")
	free.CouponCode = "FREE"
	freeOrder, err := svc.CreateOrder(ctx, "test", free)
	if err != nil || freeOrder.InvoiceStatus != "paid" || freeOrder.CheckoutURL != "" || freeOrder.SubscriptionStatus != "active" {
		t.Fatalf("free=%+v %v", freeOrder, err)
	}
	var trialPublic string
	db.Get(&trialPublic, `SELECT public_id FROM plans WHERE id=2`)
	trialQuote, err := svc.QuoteEmbedPlan(ctx, site.PublicID, trialPublic, options, "SAVE10")
	if err != nil || trialQuote.TrialDays != 7 || trialQuote.DueNow != 0 || trialQuote.Amount != 135000 {
		t.Fatalf("trial quote=%+v %v", trialQuote, err)
	}
	trial := payment.EmbedOrderRequest{PlanPublicID: trialPublic, RequestID: "trial-request", AddOns: options, CouponCode: "SAVE10"}
	trial.Customer.Name = "Trial buyer"
	trial.Customer.Email = "trial@example.invalid"
	trial.Customer.Mobile = "+628123456789"
	first, err := svc.CreateEmbedOrder(ctx, site.PublicID, "", trial)
	if err != nil || first.InvoiceID != 0 || first.CheckoutURL == "" || first.SubscriptionStatus != "pending" || first.TrialEndsAt != nil {
		t.Fatalf("trial=%+v %v", first, err)
	}
	again, err := svc.CreateEmbedOrder(ctx, site.PublicID, "", trial)
	if err != nil || again.SubscriptionID != first.SubscriptionID {
		t.Fatalf("trial retry=%+v %v", again, err)
	}
	trial.RequestID = "another-trial"
	if _, err = svc.CreateEmbedOrder(ctx, site.PublicID, "", trial); !errors.Is(err, utils.ErrConflict) {
		t.Fatalf("repeat trial=%v", err)
	}
	var count int
	db.Get(&count, `SELECT count(*) FROM invoices WHERE subscription_id=$1`, first.SubscriptionID)
	if count != 0 {
		t.Fatal("trial billed early")
	}
	// Isolate renewal to this subscription; no provider calls or SMTP sends.
	completed := finishTestTrial(t, checkout, first, gateway)
	first.TrialEndsAt = completed.TrialEndsAt
	db.Exec(`UPDATE subscriptions SET next_billing_date=NULL WHERE id<>$1`, first.SubscriptionID)
	result, err := billing.NewRenewalService(db).Run(ctx, first.TrialEndsAt.Add(time.Second), 10)
	if err != nil || result.InvoicesCreated != 1 || len(result.Failures) > 0 {
		t.Fatalf("renewal=%+v %v", result, err)
	}
	renewed, err := svc.GetOrder(ctx, "test", first.ExternalOrderID)
	if err != nil || renewed.InvoiceID == 0 || renewed.InvoiceStatus != "pending" || renewed.SubscriptionStatus != "pending" {
		t.Fatalf("renewed=%+v %v", renewed, err)
	}
	db.Get(&amount, `SELECT amount FROM invoices WHERE id=$1`, renewed.InvoiceID)
	if amount != 135000 {
		t.Fatalf("trial first bill=%v", amount)
	}
	var used int
	db.Get(&used, `SELECT used_count FROM coupons WHERE code='SAVE10'`)
	if used != 2 {
		t.Fatalf("coupon consumed again on retry: %d", used)
	}
	if err := db.Get(&count, `SELECT count(*) FROM checkout_email_deliveries`); err != nil || count != 2 {
		t.Fatalf("trial+free confirmation count=%d %v", count, err)
	}
	// Trial retries and paid replays do not enqueue duplicates; rollback leaves no job.
	rollbackTx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = rollbackTx.Exec(`UPDATE invoices SET status='paid' WHERE id=1`); err != nil {
		t.Fatal(err)
	}
	if err = billing.QueuePaymentConfirmation(ctx, rollbackTx, 1); err != nil {
		t.Fatal(err)
	}
	if err = rollbackTx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if err = db.Get(&count, `SELECT count(*) FROM checkout_email_deliveries WHERE invoice_id=1`); err != nil || count != 0 {
		t.Fatal("confirmation survived payment rollback")
	}
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err = billing.QueuePaymentConfirmation(ctx, tx, freeOrder.InvoiceID); err != nil {
		t.Fatal(err)
	}
	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}
	invoiceService := billing.NewInvoiceService(db, billingrepo.NewInvoiceRepository(db))
	previewDir := t.TempDir()
	if err = invoiceService.RunCheckoutEmailBatch(ctx, &config.Config{EmailDeliveryMode: "preview", EmailPreviewDir: previewDir}); err != nil {
		t.Fatal(err)
	}
	if err = db.Get(&count, `SELECT count(*) FROM checkout_email_deliveries WHERE status='previewed' AND attempts=1`); err != nil || count != 2 {
		t.Fatalf("confirmation worker count=%d %v", count, err)
	}
	var jobs []struct {
		ID    int    `db:"id"`
		Event string `db:"event_type"`
		Token string `db:"public_token"`
	}
	if err = db.Select(&jobs, `SELECT id,event_type,public_token FROM checkout_email_deliveries`); err != nil {
		t.Fatal(err)
	}
	for _, job := range jobs {
		body, err := os.ReadFile(filepath.Join(previewDir, fmt.Sprintf("checkout-%d.html", job.ID)))
		if err != nil {
			t.Fatal(err)
		}
		if job.Event == "trial_started" {
			if !strings.Contains(string(body), "Trial berhasil dimulai") || !strings.Contains(string(body), "Extra seats") {
				t.Fatal("trial preview missing snapshot")
			}
		} else {
			id, biz, _, err := invoiceService.PublicEmailInvoice(ctx, job.Token)
			if err != nil || id != freeOrder.InvoiceID || biz != "test" {
				t.Fatalf("receipt link %d %s %v", id, biz, err)
			}
		}
	}
	if err = invoiceService.RunCheckoutEmailBatch(ctx, &config.Config{EmailDeliveryMode: "preview", EmailPreviewDir: previewDir}); err != nil {
		t.Fatal(err)
	}
	if err = db.Get(&count, `SELECT count(*) FROM checkout_email_deliveries WHERE attempts=1`); err != nil || count != 2 {
		t.Fatal("worker replay duplicated confirmations")
	}
}
