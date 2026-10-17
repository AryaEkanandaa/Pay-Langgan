package payment_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"pay-langgan/internal/models"
	billingrepo "pay-langgan/internal/repositories/billing"
	paymentrepo "pay-langgan/internal/repositories/payment"
	payment "pay-langgan/internal/services/payment"
	"pay-langgan/internal/utils"
	provider "pay-langgan/pkg/payment"
)

type trialSessionProvider struct {
	payment.SessionProvider
	session *provider.XenditSessionResponse
}

func (p *trialSessionProvider) CreateCardSession(ctx context.Context, req provider.XenditSessionRequest) (*provider.XenditSessionResponse, error) {
	result, err := p.SessionProvider.CreateCardSession(ctx, req)
	p.session = result
	return result, err
}
func (p *trialSessionProvider) GetCardSession(ctx context.Context, id, ref string) (*provider.XenditSessionResponse, error) {
	if p.session != nil && p.session.PaymentSessionID == id && p.session.ReferenceID == ref {
		copy := *p.session
		return &copy, nil
	}
	return p.SessionProvider.GetCardSession(ctx, id, ref)
}

func finishTestTrial(t *testing.T, checkout *payment.CheckoutService, order *payment.MerchantOrderResponse, gateway *trialSessionProvider) *models.PublicCheckoutResponse {
	t.Helper()
	ctx := context.Background()
	token := strings.TrimPrefix(order.CheckoutURL, "http://localhost:5173/checkout/")
	if _, err := checkout.Start(ctx, token, models.StartCheckoutRequest{}); !errors.Is(err, utils.ErrBadRequest) {
		t.Fatalf("missing consent accepted: %v", err)
	}
	start, err := checkout.Start(ctx, token, models.StartCheckoutRequest{CardConsent: true})
	if err != nil || !start.CardRegistration {
		t.Fatalf("start: %+v %v", start, err)
	}
	gateway.session.Status = "COMPLETED"
	gateway.session.PaymentTokenID = "pt-trial-test"
	public, err := checkout.Get(ctx, token)
	if err != nil || public.CheckoutStatus != "completed" || public.TrialEndsAt == nil {
		t.Fatalf("finish: %+v %v", public, err)
	}
	return public
}

func TestTrialCardEnrollmentPostgres(t *testing.T) {
	for _, name := range []string{"xendit", "midtrans"} {
		t.Run(name, func(t *testing.T) {
			db := paymentTestDB(t)
			ctx := context.Background()
			_, err := db.Exec(`SELECT setval(pg_get_serial_sequence('customers','id'),(SELECT max(id) FROM customers)); SELECT setval(pg_get_serial_sequence('subscriptions','id'),(SELECT max(id) FROM subscriptions)); INSERT INTO platform_subscriptions(business_id,status,paid_through) VALUES('test','active',NOW()+INTERVAL '1 month'); INSERT INTO plans(id,product_id,name,price,billing_cycle,trial_days) VALUES(2,1,'Trial',100000,'monthly',7)`)
			if err != nil {
				t.Fatal(err)
			}
			mock, _ := provider.NewXenditProvider("", "test-secret", "https://api.xendit.co", true)
			gateway := &trialSessionProvider{SessionProvider: mock}
			checkout, _ := payment.NewCheckoutService(db, paymentrepo.NewCheckoutRepository(db), gateway, "http://localhost:5173", "https://paylanggan.example", time.Hour)
			runtime := &payment.BusinessGatewayRuntime{BusinessID: "test", Config: payment.BusinessGatewayInput{Provider: "midtrans", Environment: "sandbox", MidtransClientKey: "client", MidtransServerKey: "server"}}
			if name == "midtrans" {
				checkout.SetBusinessGatewayResolver(midtransTestResolver{runtime})
			}
			merchant := payment.NewMerchantIntegrationService(db, checkout, billingrepo.NewInvoiceRepository(db))
			req := payment.MerchantOrderRequest{ExternalOrderID: "trial", PlanID: 2}
			req.Customer.Name = "Buyer"
			req.Customer.Email = "buyer@example.invalid"
			req.Customer.Mobile = "+628123456789"
			order, err := merchant.CreateOrder(ctx, "test", req)
			if err != nil {
				t.Fatal(err)
			}
			if order.SubscriptionStatus != "pending" || order.TrialEndsAt != nil || order.InvoiceID != 0 {
				t.Fatalf("trial active before enrollment: %+v", order)
			}
			retry, err := merchant.CreateOrder(ctx, "test", req)
			if err != nil || retry.CheckoutURL != order.CheckoutURL {
				t.Fatalf("idempotency: %+v %v", retry, err)
			}
			var n int
			if err = db.Get(&n, `SELECT count(*) FROM checkout_email_deliveries`); err != nil || n != 0 {
				t.Fatalf("premature email: %d %v", n, err)
			}
			token := strings.TrimPrefix(order.CheckoutURL, "http://localhost:5173/checkout/")
			if name == "xendit" {
				if _, err = checkout.RegisterMidtransTrialCard(ctx, token, payment.MidtransTrialRegistration{SavedTokenID: "registered-token-1234567890", TransactionID: "50e7c4ac-772e-43fa-94fb-52559bce4fc0"}); !errors.Is(err, utils.ErrConflict) {
					t.Fatalf("wrong provider accepted: %v", err)
				}
				finishTestTrial(t, checkout, order, gateway)
				payload := []byte(fmt.Sprintf(`{"event":"payment_session.completed","data":{"payment_session_id":%q,"reference_id":%q}}`, gateway.session.PaymentSessionID, gateway.session.ReferenceID))
				if _, err = checkout.ProcessWebhook(ctx, "invalid", payload); !errors.Is(err, utils.ErrUnauthorized) {
					t.Fatalf("bad callback accepted: %v", err)
				}
				duplicate, e := checkout.ProcessWebhook(ctx, "test-secret", payload)
				if e != nil || !duplicate {
					t.Fatalf("replay: %v %v", duplicate, e)
				}
			} else {
				receipt := payment.MidtransTrialRegistration{SavedTokenID: "registered-token-1234567890", TransactionID: "50e7c4ac-772e-43fa-94fb-52559bce4fc0"}
				if _, err = checkout.RegisterMidtransTrialCard(ctx, token, receipt); !errors.Is(err, utils.ErrConflict) {
					t.Fatalf("registration without consent: %v", err)
				}
				if _, err = checkout.Start(ctx, token, models.StartCheckoutRequest{CardConsent: true}); err != nil {
					t.Fatal(err)
				}
				runtime.Config.MidtransServerKey = "different-account"
				if _, err = checkout.RegisterMidtransTrialCard(ctx, token, receipt); !errors.Is(err, utils.ErrConflict) {
					t.Fatalf("account mismatch accepted: %v", err)
				}
				runtime.Config.MidtransServerKey = "server"
				bad := receipt
				bad.SavedTokenID = "4811111111111114"
				if _, err = checkout.RegisterMidtransTrialCard(ctx, token, bad); !errors.Is(err, utils.ErrBadRequest) {
					t.Fatalf("raw PAN accepted: %v", err)
				}
				for i := 0; i < 2; i++ {
					public, e := checkout.RegisterMidtransTrialCard(ctx, token, receipt)
					if e != nil || public.CheckoutStatus != "completed" {
						t.Fatalf("registration: %+v %v", public, e)
					}
				}
				var level string
				if err = db.Get(&level, `SELECT verification_level FROM trial_card_sessions WHERE subscription_id=$1`, order.SubscriptionID); err != nil || level != "registered" {
					t.Fatalf("registration mislabeled: %s %v", level, err)
				}
			}
			final, err := merchant.GetOrder(ctx, "test", order.ExternalOrderID)
			if err != nil || final.SubscriptionStatus != "trial" || final.TrialEndsAt == nil || time.Until(*final.TrialEndsAt) < 6*24*time.Hour {
				t.Fatalf("trial timing: %+v %v", final, err)
			}
			if err = db.Get(&n, `SELECT count(*) FROM invoices WHERE subscription_id=$1`, order.SubscriptionID); err != nil || n != 0 {
				t.Fatalf("enrollment billed: %d %v", n, err)
			}
			if err = db.Get(&n, `SELECT count(*) FROM subscription_autopay a JOIN payment_methods pm ON pm.id=a.payment_method_id WHERE a.subscription_id=$1 AND a.enabled=TRUE AND pm.provider=$2`, order.SubscriptionID, name); err != nil || n != 1 {
				t.Fatalf("enrollment did not enable consented %s debit: %d %v", name, n, err)
			}
			if err = db.Get(&n, `SELECT count(*) FROM checkout_email_deliveries WHERE subscription_id=$1`, order.SubscriptionID); err != nil || n != 1 {
				t.Fatalf("confirmation duplicated: %d %v", n, err)
			}
		})
	}
}

func TestAbandonedTrialReleasesCoupon(t *testing.T) {
	db := paymentTestDB(t)
	_, err := db.Exec(`SELECT setval(pg_get_serial_sequence('customers','id'),(SELECT max(id) FROM customers)); SELECT setval(pg_get_serial_sequence('subscriptions','id'),(SELECT max(id) FROM subscriptions)); INSERT INTO platform_subscriptions(business_id,status,paid_through) VALUES('test','active',NOW()+INTERVAL '1 month'); INSERT INTO plans(id,product_id,name,price,billing_cycle,trial_days) VALUES(2,1,'Trial',100000,'monthly',7); INSERT INTO coupons(id,business_id,code,discount_type,discount_value,max_usage) VALUES(1,'test','ONE','fixed',10000,1)`)
	if err != nil {
		t.Fatal(err)
	}
	mock, _ := provider.NewXenditProvider("", "test-secret", "https://api.xendit.co", true)
	checkout, _ := payment.NewCheckoutService(db, paymentrepo.NewCheckoutRepository(db), mock, "http://localhost:5173", "https://paylanggan.example", time.Hour)
	merchant := payment.NewMerchantIntegrationService(db, checkout, billingrepo.NewInvoiceRepository(db))
	req := payment.MerchantOrderRequest{ExternalOrderID: "abandoned", PlanID: 2, CouponCode: "ONE"}
	req.Customer.Name = "Buyer"
	req.Customer.Email = "buyer@example.invalid"
	req.Customer.Mobile = "+628123456789"
	first, err := merchant.CreateOrder(context.Background(), "test", req)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`UPDATE trial_card_sessions SET expires_at=NOW()-INTERVAL '1 minute' WHERE subscription_id=$1`, first.SubscriptionID); err != nil {
		t.Fatal(err)
	}
	req.ExternalOrderID = "next"
	req.Customer.Email = "next@example.invalid"
	if _, err = merchant.CreateOrder(context.Background(), "test", req); err != nil {
		t.Fatal("coupon was not released:", err)
	}
	var used int
	if err = db.Get(&used, `SELECT used_count FROM coupons WHERE id=1`); err != nil || used != 1 {
		t.Fatalf("coupon count=%d %v", used, err)
	}
	var status string
	if err = db.Get(&status, `SELECT status FROM subscriptions WHERE id=$1`, first.SubscriptionID); err != nil || status != "cancelled" {
		t.Fatalf("abandoned trial status=%s %v", status, err)
	}
}

func TestTrialRecoversAfterCustomerClosesCheckout(t *testing.T) {
	db := paymentTestDB(t)
	ctx := context.Background()
	_, err := db.Exec(`SELECT setval(pg_get_serial_sequence('customers','id'),(SELECT max(id) FROM customers)); SELECT setval(pg_get_serial_sequence('subscriptions','id'),(SELECT max(id) FROM subscriptions)); INSERT INTO platform_subscriptions(business_id,status,paid_through) VALUES('test','active',NOW()+INTERVAL '1 month'); INSERT INTO plans(id,product_id,name,price,billing_cycle,trial_days) VALUES(2,1,'Trial',100000,'monthly',7)`)
	if err != nil {
		t.Fatal(err)
	}
	mock, _ := provider.NewXenditProvider("", "test-secret", "https://api.xendit.co", true)
	gateway := &trialSessionProvider{SessionProvider: mock}
	checkout, _ := payment.NewCheckoutService(db, paymentrepo.NewCheckoutRepository(db), gateway, "http://localhost:5173", "https://paylanggan.example", time.Hour)
	merchant := payment.NewMerchantIntegrationService(db, checkout, billingrepo.NewInvoiceRepository(db))
	req := payment.MerchantOrderRequest{ExternalOrderID: "closed", PlanID: 2}
	req.Customer.Name = "Buyer"
	req.Customer.Email = "buyer@example.invalid"
	req.Customer.Mobile = "+628123456789"
	order, err := merchant.CreateOrder(ctx, "test", req)
	if err != nil {
		t.Fatal(err)
	}
	token := strings.TrimPrefix(order.CheckoutURL, "http://localhost:5173/checkout/")
	if _, err = checkout.Start(ctx, token, models.StartCheckoutRequest{CardConsent: true}); err != nil {
		t.Fatal(err)
	}
	gateway.session.Status = "COMPLETED"
	gateway.session.PaymentTokenID = "pt-closed"
	if _, err = checkout.RunMaintenance(ctx, 10); err != nil {
		t.Fatal(err)
	}
	final, err := merchant.GetOrder(ctx, "test", order.ExternalOrderID)
	if err != nil || final.SubscriptionStatus != "trial" {
		t.Fatalf("closed checkout not recovered: %+v %v", final, err)
	}
}
