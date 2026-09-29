package payment

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestXenditRecurringRequestAndLookup(t *testing.T) {
	posts, gets := 0, 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key, _, ok := r.BasicAuth()
		if !ok || key != "test-key" || r.Header.Get("api-version") != "2024-11-11" {
			t.Error("missing Xendit authentication/version")
		}
		if r.Method == http.MethodPost {
			posts++
			if r.URL.Path != "/v3/payment_requests" {
				t.Error("wrong create endpoint")
			}
			var payload struct {
				Type       string `json:"type"`
				Token      string `json:"payment_token_id"`
				Amount     int64  `json:"request_amount"`
				Properties struct {
					Type          string `json:"card_on_file_type"`
					Sequence      string `json:"transaction_sequence"`
					Configuration struct {
						Frequency int    `json:"recurring_frequency"`
						Expiry    string `json:"recurring_expiry"`
					} `json:"recurring_configuration"`
				} `json:"channel_properties"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Error(err)
			}
			if payload.Type != "PAY" || payload.Token != "pt-test" || payload.Amount != 150000 || payload.Properties.Type != "RECURRING" || payload.Properties.Sequence != "SUBSEQUENT" || payload.Properties.Configuration.Frequency != 28 || payload.Properties.Configuration.Expiry == "" {
				t.Error("incorrect recurring payment payload")
			}
		} else {
			gets++
			if r.Method != http.MethodGet || r.URL.Path != "/v3/payment_requests/pr-test" {
				t.Error("wrong lookup endpoint")
			}
		}
		_ = json.NewEncoder(w).Encode(RecurringChargeResponse{PaymentRequestID: "pr-test", ReferenceID: "ref-test", Currency: "IDR", Amount: 150000, Status: "SUCCEEDED", PaymentID: "py-test"})
	}))
	defer server.Close()
	p, err := NewXenditProvider("test-key", "callback", server.URL, false)
	if err != nil {
		t.Fatal(err)
	}
	p.httpClient = server.Client()
	r := RecurringChargeRequest{ReferenceID: "ref-test", PaymentTokenID: "pt-test", Amount: 150000, ReturnURL: "https://example.com", FrequencyDays: 28, ConsentExpiresAt: time.Now().AddDate(1, 0, 0)}
	response, err := p.CreateRecurringCharge(context.Background(), r)
	if err != nil || response.PaymentID != "py-test" {
		t.Fatal("create charge failed", err)
	}
	response, err = p.GetRecurringCharge(context.Background(), "pr-test")
	if err != nil || response.Status != "SUCCEEDED" {
		t.Fatal("lookup charge failed", err)
	}
	r.ConsentExpiresAt = time.Now().Add(-time.Hour)
	if _, err := p.CreateRecurringCharge(context.Background(), r); err == nil {
		t.Fatal("expired authorization accepted")
	}
	if posts != 1 || gets != 1 {
		t.Fatal("unexpected provider retry")
	}
}
