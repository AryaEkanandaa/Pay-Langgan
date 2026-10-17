package payment_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"pay-langgan/internal/models"
	billingrepo "pay-langgan/internal/repositories/billing"
	paymentrepo "pay-langgan/internal/repositories/payment"
	paymentservice "pay-langgan/internal/services/payment"
	"pay-langgan/internal/utils"
	provider "pay-langgan/pkg/payment"
)

func TestCancelMerchantCheckoutReleasesPendingOrder(t *testing.T) {
	db := paymentTestDB(t)
	_, err := db.Exec(`SELECT setval(pg_get_serial_sequence('customers','id'),(SELECT max(id) FROM customers));
	 SELECT setval(pg_get_serial_sequence('subscriptions','id'),(SELECT max(id) FROM subscriptions));
	 SELECT setval(pg_get_serial_sequence('invoices','id'),(SELECT max(id) FROM invoices));
	 INSERT INTO platform_subscriptions(business_id,status,paid_through) VALUES('test','active',NOW()+INTERVAL '1 month')`)
	if err != nil {
		t.Fatal(err)
	}
	mock, _ := provider.NewXenditProvider("", "callback-secret", "https://api.xendit.co", true)
	checkout, err := paymentservice.NewCheckoutService(db, paymentrepo.NewCheckoutRepository(db), mock, "http://localhost:5173", "https://paylanggan.example", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	merchant := paymentservice.NewMerchantIntegrationService(db, checkout, billingrepo.NewInvoiceRepository(db))
	ctx := context.Background()
	req := paymentservice.MerchantOrderRequest{ExternalOrderID: "first", PlanID: 1}
	req.Customer.Name, req.Customer.Email, req.Customer.Mobile = "Buyer", "buyer@example.invalid", "+628123456789"
	order, err := merchant.CreateOrder(ctx, "test", req)
	if err != nil {
		t.Fatal(err)
	}
	token := strings.TrimPrefix(order.CheckoutURL, "http://localhost:5173/checkout/")
	started, err := checkout.Start(ctx, token, models.StartCheckoutRequest{Name: req.Customer.Name, Email: req.Customer.Email, Mobile: req.Customer.Mobile})
	if err != nil {
		t.Fatal(err)
	}
	cancelled, err := checkout.Cancel(ctx, token)
	if err != nil || cancelled.CheckoutStatus != "cancelled" {
		t.Fatalf("cancelled=%+v err=%v", cancelled, err)
	}
	var reference string
	if err = db.Get(&reference, `SELECT reference_id FROM checkout_sessions WHERE invoice_id=$1`, order.InvoiceID); err != nil {
		t.Fatal(err)
	}
	providerSession, err := mock.GetCardSession(ctx, started.PaymentSessionID, reference)
	if err != nil || providerSession.Status != "CANCELED" {
		t.Fatalf("gateway session=%+v err=%v", providerSession, err)
	}
	if _, err = checkout.Start(ctx, token, models.StartCheckoutRequest{}); !errors.Is(err, utils.ErrConflict) {
		t.Fatalf("cancelled checkout started: %v", err)
	}
	var invoiceStatus, subscriptionStatus string
	if err = db.QueryRow(`SELECT i.status,s.status FROM invoices i JOIN subscriptions s ON s.id=i.subscription_id WHERE i.id=$1`, order.InvoiceID).Scan(&invoiceStatus, &subscriptionStatus); err != nil {
		t.Fatal(err)
	}
	if invoiceStatus != "cancelled" || subscriptionStatus != "cancelled" {
		t.Fatalf("invoice=%s subscription=%s", invoiceStatus, subscriptionStatus)
	}
	req.ExternalOrderID = "second"
	if _, err = merchant.CreateOrder(ctx, "test", req); err != nil {
		t.Fatalf("new checkout after cancellation: %v", err)
	}
}

func TestCancelTrialCardSessionBeforeVerification(t *testing.T) {
	db := paymentTestDB(t)
	_, err := db.Exec(`SELECT setval(pg_get_serial_sequence('customers','id'),(SELECT max(id) FROM customers));
	 SELECT setval(pg_get_serial_sequence('subscriptions','id'),(SELECT max(id) FROM subscriptions));
	 INSERT INTO platform_subscriptions(business_id,status,paid_through) VALUES('test','active',NOW()+INTERVAL '1 month');
	 INSERT INTO plans(id,product_id,name,price,billing_cycle,trial_days) VALUES(2,1,'Trial',100000,'monthly',7);
	 INSERT INTO coupons(id,business_id,code,discount_type,discount_value,max_usage) VALUES(1,'test','ONE','fixed',10000,1)`)
	if err != nil {
		t.Fatal(err)
	}
	mock, _ := provider.NewXenditProvider("", "callback-secret", "https://api.xendit.co", true)
	checkout, err := paymentservice.NewCheckoutService(db, paymentrepo.NewCheckoutRepository(db), mock, "http://localhost:5173", "https://paylanggan.example", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	merchant := paymentservice.NewMerchantIntegrationService(db, checkout, billingrepo.NewInvoiceRepository(db))
	ctx := context.Background()
	req := paymentservice.MerchantOrderRequest{ExternalOrderID: "trial-first", PlanID: 2, CouponCode: "ONE"}
	req.Customer.Name, req.Customer.Email, req.Customer.Mobile = "Buyer", "buyer@example.invalid", "+628123456789"
	order, err := merchant.CreateOrder(ctx, "test", req)
	if err != nil {
		t.Fatal(err)
	}
	token := strings.TrimPrefix(order.CheckoutURL, "http://localhost:5173/checkout/")
	started, err := checkout.Start(ctx, token, models.StartCheckoutRequest{CardConsent: true})
	if err != nil {
		t.Fatal(err)
	}
	cancelled, err := checkout.Cancel(ctx, token)
	if err != nil || cancelled.CheckoutStatus != "cancelled" {
		t.Fatalf("cancelled=%+v err=%v", cancelled, err)
	}
	var reference string
	if err = db.Get(&reference, `SELECT reference_id FROM trial_card_sessions WHERE subscription_id=$1`, order.SubscriptionID); err != nil {
		t.Fatal(err)
	}
	providerSession, err := mock.GetCardSession(ctx, started.PaymentSessionID, reference)
	if err != nil || providerSession.Status != "CANCELED" {
		t.Fatalf("gateway session=%+v err=%v", providerSession, err)
	}
	var subscriptionStatus string
	if err = db.Get(&subscriptionStatus, `SELECT status FROM subscriptions WHERE id=$1`, order.SubscriptionID); err != nil || subscriptionStatus != "cancelled" {
		t.Fatalf("subscription=%s err=%v", subscriptionStatus, err)
	}
	var used int
	if err = db.Get(&used, `SELECT used_count FROM coupons WHERE id=1`); err != nil || used != 0 {
		t.Fatalf("coupon usage=%d err=%v", used, err)
	}
	req.ExternalOrderID = "trial-second"
	if _, err = merchant.CreateOrder(ctx, "test", req); err != nil {
		t.Fatalf("new trial after cancel: %v", err)
	}
}

func TestCancelMidtransCheckoutBeforeCardCharge(t *testing.T) {
	db := paymentTestDB(t)
	_, err := db.Exec(`SELECT setval(pg_get_serial_sequence('customers','id'),(SELECT max(id) FROM customers));
	 SELECT setval(pg_get_serial_sequence('subscriptions','id'),(SELECT max(id) FROM subscriptions));
	 SELECT setval(pg_get_serial_sequence('invoices','id'),(SELECT max(id) FROM invoices));
	 INSERT INTO platform_subscriptions(business_id,status,paid_through) VALUES('test','active',NOW()+INTERVAL '1 month')`)
	if err != nil {
		t.Fatal(err)
	}
	mock, _ := provider.NewXenditProvider("", "callback-secret", "https://api.xendit.co", true)
	checkout, err := paymentservice.NewCheckoutService(db, paymentrepo.NewCheckoutRepository(db), mock, "http://localhost:5173", "https://paylanggan.example", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	checkout.SetBusinessGatewayResolver(midtransTestResolver{&paymentservice.BusinessGatewayRuntime{BusinessID: "test", Config: paymentservice.BusinessGatewayInput{Provider: "midtrans", Environment: "sandbox", MidtransClientKey: "client", MidtransServerKey: "server"}}})
	merchant := paymentservice.NewMerchantIntegrationService(db, checkout, billingrepo.NewInvoiceRepository(db))
	ctx := context.Background()
	req := paymentservice.MerchantOrderRequest{ExternalOrderID: "midtrans-cancel", PlanID: 1}
	req.Customer.Name, req.Customer.Email, req.Customer.Mobile = "Buyer", "buyer@example.invalid", "+628123456789"
	order, err := merchant.CreateOrder(ctx, "test", req)
	if err != nil {
		t.Fatal(err)
	}
	token := strings.TrimPrefix(order.CheckoutURL, "http://localhost:5173/checkout/")
	started, err := checkout.Start(ctx, token, models.StartCheckoutRequest{Name: req.Customer.Name, Email: req.Customer.Email, Mobile: req.Customer.Mobile})
	if err != nil || started.Mode != "CORE" {
		t.Fatalf("start=%+v err=%v", started, err)
	}
	cancelled, err := checkout.Cancel(ctx, token)
	if err != nil || cancelled.CheckoutStatus != "cancelled" {
		t.Fatalf("cancelled=%+v err=%v", cancelled, err)
	}
	if _, err = checkout.Start(ctx, token, models.StartCheckoutRequest{}); !errors.Is(err, utils.ErrConflict) {
		t.Fatalf("cancelled Midtrans checkout started: %v", err)
	}
}
