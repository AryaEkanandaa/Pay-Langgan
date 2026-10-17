package payment

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMidtransRecurringChargeAndReconciliation(t *testing.T) {
	var status = "settlement"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if key, password, ok := r.BasicAuth(); !ok || key != "tenant-server" || password != "" {
			t.Error("wrong Midtrans tenant authentication")
		}
		if r.Method == http.MethodPost {
			if r.URL.Path != "/v2/charge" {
				t.Errorf("unexpected charge endpoint: %s", r.URL.Path)
			}
			var request struct {
				PaymentType string `json:"payment_type"`
				Transaction struct {
					OrderID string `json:"order_id"`
					Amount  int64  `json:"gross_amount"`
				} `json:"transaction_details"`
				Card struct {
					TokenID string `json:"token_id"`
					Auth    *bool  `json:"authentication"`
				} `json:"credit_card"`
			}
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Fatal(err)
			}
			if request.PaymentType != "credit_card" || request.Transaction.OrderID != "autopay_123" ||
				request.Transaction.Amount != 125000 || request.Card.TokenID != "saved-card-token" || request.Card.Auth != nil {
				t.Errorf("unexpected recurring request: %+v", request)
			}
		} else if r.Method != http.MethodGet || r.URL.Path != "/v2/autopay_123/status" {
			t.Errorf("unexpected status endpoint: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status_code":"200","transaction_id":"tx-456","order_id":"autopay_123","gross_amount":"125000.00","currency":"IDR","payment_type":"credit_card","transaction_status":"` + status + `","fraud_status":"accept"}`))
	}))
	defer server.Close()
	client := NewMidtransProvider("tenant-server", "tenant-client", server.URL, false)
	response, err := client.CreateRecurringCharge(context.Background(), RecurringChargeRequest{
		ReferenceID: "autopay_123", PaymentTokenID: "saved-card-token", Amount: 125000,
	})
	if err != nil || response.Status != "SUCCEEDED" || response.PaymentRequestID != "autopay_123" || response.PaymentID != "tx-456" {
		t.Fatalf("recurring charge response: %+v, %v", response, err)
	}
	response, err = client.GetRecurringCharge(context.Background(), "autopay_123")
	if err != nil || response.Status != "SUCCEEDED" || response.ReferenceID != "autopay_123" {
		t.Fatalf("reconciliation response: %+v, %v", response, err)
	}
	status = "deny"
	response, err = client.GetRecurringCharge(context.Background(), "autopay_123")
	if err != nil || response.Status != "FAILED" || response.FailureCode != "CARD_DECLINED" {
		t.Fatalf("failed transaction mapping: %+v, %v", response, err)
	}
}

func TestMidtransRecurringRejectsMismatchedAmount(t *testing.T) {
	transaction := &MidtransCardTransaction{
		TransactionID: "tx-1", OrderID: "autopay_1", GrossAmount: "120000.00", PaymentType: "credit_card", TransactionStatus: "settlement",
	}
	if _, err := midtransRecurringResponse(transaction, "autopay_1", 125000); err == nil {
		t.Fatal("accepted a Midtrans response with the wrong amount")
	}
}
