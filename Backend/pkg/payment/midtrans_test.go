package payment

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateSnapTransactionUsesTenantServerKeyAndAmount(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/snap/v1/transactions" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		user, password, ok := r.BasicAuth()
		if !ok || user != "tenant-server-key" || password != "" {
			t.Fatalf("unexpected basic auth user=%q password=%q", user, password)
		}
		var payload struct {
			EnabledPayments    []string `json:"enabled_payments"`
			TransactionDetails struct {
				OrderID     string `json:"order_id"`
				GrossAmount int64  `json:"gross_amount"`
			} `json:"transaction_details"`
			CreditCard struct {
				Secure bool `json:"secure"`
			} `json:"credit_card"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		if len(payload.EnabledPayments) != 1 || payload.EnabledPayments[0] != "credit_card" {
			t.Fatalf("only credit card must be enabled: %v", payload.EnabledPayments)
		}
		if payload.TransactionDetails.OrderID != "invoice_15_abc123" || payload.TransactionDetails.GrossAmount != 100000 || !payload.CreditCard.Secure {
			t.Fatalf("unexpected transaction payload: %+v", payload)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"token":"snap-token","redirect_url":"https://app.sandbox.midtrans.com/snap/v2/vtweb/snap-token"}`))
	}))
	defer server.Close()

	client := NewMidtransProvider("tenant-server-key", "public-client-key", server.URL, false)
	response, err := client.CreateSnapTransaction(context.Background(), MidtransSnapRequest{
		OrderID: "invoice_15_abc123", Amount: 100000, FirstName: "Arya", Email: "arya@example.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	if response.Token != "snap-token" {
		t.Fatalf("token = %q", response.Token)
	}
}

func TestVerifyMidtransNotificationSignature(t *testing.T) {
	const orderID, statusCode, grossAmount, serverKey = "ORDER-123", "200", "100000.00", "tenant-server-key"
	sum := sha512.Sum512([]byte(orderID + statusCode + grossAmount + serverKey))
	signature := hex.EncodeToString(sum[:])
	if !VerifyMidtransNotificationSignature(orderID, statusCode, grossAmount, serverKey, signature) {
		t.Fatal("expected valid notification signature")
	}
	if VerifyMidtransNotificationSignature(orderID, "201", grossAmount, serverKey, signature) {
		t.Fatal("accepted a signature with a modified status code")
	}
}
