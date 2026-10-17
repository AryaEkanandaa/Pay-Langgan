package payment

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMidtransCoreCardRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key, password, ok := r.BasicAuth()
		if !ok || key != "tenant-key" || password != "" {
			t.Error("wrong tenant credentials")
		}
		if r.Method == http.MethodPost {
			if r.URL.Path != "/v2/charge" {
				t.Error("wrong charge endpoint")
			}
			var request struct {
				PaymentType string `json:"payment_type"`
				Transaction struct {
					Order  string `json:"order_id"`
					Amount int64  `json:"gross_amount"`
				} `json:"transaction_details"`
				Card map[string]any `json:"credit_card"`
			}
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Fatal(err)
			}
			if request.PaymentType != "credit_card" || request.Transaction.Order != "invoice_1_abc" || request.Transaction.Amount != 100000 ||
				request.Card["token_id"] != "card-token" || request.Card["authentication"] != true || len(request.Card) != 2 {
				t.Errorf("incorrect token-based charge: %+v", request)
			}
		} else if r.Method != http.MethodGet || r.URL.Path != "/v2/invoice_1_abc/status" {
			t.Error("wrong status endpoint")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status_code":"201","transaction_id":"transaction-1","order_id":"invoice_1_abc","gross_amount":"100000.00","currency":"IDR","payment_type":"credit_card","transaction_status":"pending","merchant_id":"merchant-1","redirect_url":"https://api.sandbox.midtrans.com/v2/token/authentication"}`))
	}))
	defer server.Close()
	client := NewMidtransProvider("tenant-key", "client-key", server.URL, false)
	result, err := client.ChargeCard(context.Background(), MidtransCardRequest{OrderID: "invoice_1_abc", Amount: 100000, TokenID: "card-token"})
	if err != nil || result.TransactionID != "transaction-1" {
		t.Fatalf("charge failed: %v", err)
	}
	if _, err = client.GetCardTransaction(context.Background(), "invoice_1_abc"); err != nil {
		t.Fatal(err)
	}
}

func TestMidtransAuthenticationURL(t *testing.T) {
	if !ValidMidtransAuthenticationURL("https://card.midtrans.com/v2/token/authentication") {
		t.Error("current Midtrans 3DS domain rejected")
	}
	for _, value := range []string{"http://api.midtrans.com/path", "https://api.midtrans.com.evil.invalid/path", "https://evil.invalid/path", "https://user@api.midtrans.com/path", "https://api.midtrans.com:8443/path"} {
		if ValidMidtransAuthenticationURL(value) {
			t.Errorf("unsafe URL accepted: %s", value)
		}
	}
	if !ValidMidtransAuthenticationURL("https://api.sandbox.veritrans.co.id/v2/token/authentication") {
		t.Error("valid authentication URL rejected")
	}
}
