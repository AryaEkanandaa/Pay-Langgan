package payment

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHostedSessionPayAndSavePayload(t *testing.T) {
	var calls int
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path != "/sessions" || r.Method != http.MethodPost {
			t.Errorf("wrong endpoint: %s %s", r.Method, r.URL.Path)
		}
		var payload struct {
			Reference   string   `json:"reference_id"`
			SessionType string   `json:"session_type"`
			Mode        string   `json:"mode"`
			Save        string   `json:"allow_save_payment_method"`
			Channels    []string `json:"allowed_payment_channels"`
			Customer    struct {
				Reference string `json:"reference_id"`
			} `json:"customer"`
			Properties struct {
				Cards struct {
					Type          string `json:"card_on_file_type"`
					Configuration struct {
						Frequency int `json:"recurring_frequency"`
					} `json:"recurring_configuration"`
				} `json:"cards"`
			} `json:"channel_properties"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Error(err)
		}
		if payload.Reference != "platform_test" || payload.Customer.Reference != payload.Reference || payload.SessionType != "PAY" || payload.Mode != "PAYMENT_LINK" || payload.Save != "FORCED" || len(payload.Channels) != 1 || payload.Channels[0] != "CARDS" || payload.Properties.Cards.Type != "RECURRING" || payload.Properties.Cards.Configuration.Frequency != 30 {
			t.Errorf("invalid hosted session payload: %+v", payload)
		}
		_ = json.NewEncoder(w).Encode(XenditSessionResponse{PaymentSessionID: "ps-test", ReferenceID: "platform_test", Amount: 300000, Currency: "IDR", Mode: "PAYMENT_LINK", Status: "ACTIVE", PaymentLinkURL: "https://xen.to/test"})
	}))
	defer server.Close()
	p, err := NewXenditProvider("test-key", "callback", server.URL, false)
	if err != nil {
		t.Fatal(err)
	}
	p.httpClient = server.Client()
	_, err = p.CreateHostedSession(context.Background(), HostedSessionRequest{ReferenceID: "platform_test", BusinessID: "business", Amount: 300000, PayerName: "Owner", PayerEmail: "owner@example.com", SuccessURL: "https://example.com/success", CancelURL: "https://example.com/cancel", SavePaymentMethod: true})
	if err != nil || calls != 1 {
		t.Fatalf("hosted session failed: %v, calls=%d", err, calls)
	}
}
