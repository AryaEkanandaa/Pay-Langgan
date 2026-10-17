package payment_test

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"testing"
	"time"

	paymentrepo "pay-langgan/internal/repositories/payment"
	paymentservice "pay-langgan/internal/services/payment"
	"pay-langgan/internal/utils"
	provider "pay-langgan/pkg/payment"
)

type midtransTestResolver struct {
	runtime *paymentservice.BusinessGatewayRuntime
}

func (r midtransTestResolver) LoadForBusiness(context.Context, string) (*paymentservice.BusinessGatewayRuntime, error) {
	return r.runtime, nil
}
func (r midtransTestResolver) LoadForWebhook(context.Context, string) (*paymentservice.BusinessGatewayRuntime, error) {
	return r.runtime, nil
}

func TestMidtransWebhookPostgres(t *testing.T) {
	db := paymentTestDB(t)
	if _, err := db.Exec(`UPDATE subscriptions SET status='pending' WHERE id=1`); err != nil {
		t.Fatal(err)
	}
	xendit, err := provider.NewXenditProvider("", "unused", "https://api.xendit.co", true)
	if err != nil {
		t.Fatal(err)
	}
	service, err := paymentservice.NewCheckoutService(db, paymentrepo.NewCheckoutRepository(db), xendit,
		"http://localhost:5173", "https://paylanggan.example", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	service.SetBusinessGatewayResolver(midtransTestResolver{&paymentservice.BusinessGatewayRuntime{
		BusinessID: "test", Config: paymentservice.BusinessGatewayInput{
			Provider: "midtrans", Environment: "sandbox", MidtransMerchantID: "merchant-test", MidtransServerKey: "server-test",
		},
	}})
	ctx := context.Background()
	if _, err := db.Exec(`INSERT INTO business_payment_gateways(business_id,provider,environment,encrypted_credentials,webhook_id) VALUES('test','midtrans','sandbox','\x00','test-hook')`); err != nil {
		t.Fatal(err)
	}
	notification := func(order, amount, merchant, key string) []byte {
		digest := sha512.Sum512([]byte(order + "200" + amount + key))
		payload, marshalErr := json.Marshal(map[string]string{
			"order_id": order, "gross_amount": amount, "status_code": "200", "merchant_id": merchant,
			"transaction_id": "transaction-test", "transaction_status": "settlement", "signature_key": hex.EncodeToString(digest[:]),
		})
		if marshalErr != nil {
			t.Fatal(marshalErr)
		}
		return payload
	}
	for _, credentials := range [][2]string{{"merchant-test", "wrong-key"}, {"wrong-merchant", "server-test"}} {
		_, err := service.ProcessBusinessGatewayWebhook(ctx, "test-hook", "", notification("dashboard-test", "150000.00", credentials[0], credentials[1]))
		if !errors.Is(err, utils.ErrUnauthorized) {
			t.Fatalf("invalid credentials accepted: %v", err)
		}
	}
	if _, err := service.ProcessBusinessGatewayWebhook(ctx, "test-hook", "", notification("dashboard-test", "150000.00", "merchant-test", "server-test")); err != nil {
		t.Fatalf("authenticated external order should be acknowledged: %v", err)
	}
	logs, err := service.ListWebhookLogs(ctx, "test", "midtrans", "failed", 1)
	if err != nil || logs.Total != 2 || len(logs.Items) != 2 || logs.Items[0].ErrorCode != "authentication_failed" || logs.Items[0].HTTPStatus != 401 {
		t.Fatalf("failed deliveries missing: %+v %v", logs, err)
	}
	other, err := service.ListWebhookLogs(ctx, "another-business", "", "", 1)
	if err != nil || other.Total != 0 {
		t.Fatal("webhook logs crossed tenant boundary", err)
	}
	if _, err := service.ListWebhookLogs(ctx, "test", "invalid", "", 1); !errors.Is(err, utils.ErrBadRequest) {
		t.Fatal("invalid filter accepted")
	}
	var status string
	var count int
	if err := db.Get(&status, `SELECT status FROM invoices WHERE id=1`); err != nil || status != "pending" {
		t.Fatalf("invoice changed: %s %v", status, err)
	}
	for _, table := range []string{"payments", "webhook_events"} {
		if err := db.Get(&count, "SELECT count(*) FROM "+table); err != nil || count != 0 {
			t.Fatalf("unexpected %s effects: %d %v", table, count, err)
		}
	}
	if _, err := service.CreateLink(ctx, 1, "test"); err != nil {
		t.Fatal(err)
	}
	var reference string
	if err := db.Get(&reference, `SELECT reference_id FROM checkout_sessions WHERE invoice_id=1`); err != nil {
		t.Fatal(err)
	}
	if _, err := service.ProcessBusinessGatewayWebhook(ctx, "test-hook", "", notification(reference, "1.00", "merchant-test", "server-test")); !errors.Is(err, utils.ErrBadRequest) {
		t.Fatalf("mismatched amount accepted: %v", err)
	}
	payload := notification(reference, "150000.00", "merchant-test", "server-test")
	if duplicate, err := service.ProcessBusinessGatewayWebhook(ctx, "test-hook", "", payload); err != nil || duplicate {
		t.Fatalf("valid payment failed: %v %v", duplicate, err)
	}
	if duplicate, err := service.ProcessBusinessGatewayWebhook(ctx, "test-hook", "", payload); err != nil || !duplicate {
		t.Fatalf("duplicate not handled: %v %v", duplicate, err)
	}
	if err := db.Get(&status, `SELECT status FROM invoices WHERE id=1`); err != nil || status != "paid" {
		t.Fatalf("invoice not paid: %s %v", status, err)
	}
	if err := db.Get(&count, `SELECT count(*) FROM payments WHERE status='success'`); err != nil || count != 1 {
		t.Fatalf("payment count: %d %v", count, err)
	}
	if err := db.Get(&status, `SELECT status FROM subscriptions WHERE id=1`); err != nil || status != "active" {
		t.Fatalf("Midtrans subscription=%s %v", status, err)
	}
	if err := db.Get(&count, `SELECT count(*) FROM checkout_email_deliveries WHERE invoice_id=1 AND event_type='payment_succeeded'`); err != nil || count != 1 {
		t.Fatalf("Midtrans confirmation count=%d %v", count, err)
	}
	if _, err := db.Exec(`DROP TABLE checkout_sessions CASCADE`); err != nil {
		t.Fatal(err)
	}
	if _, err := service.ProcessBusinessGatewayWebhook(ctx, "test-hook", "", notification("dashboard-test", "150000.00", "merchant-test", "server-test")); err == nil {
		t.Fatal("database failure was acknowledged")
	}
}
