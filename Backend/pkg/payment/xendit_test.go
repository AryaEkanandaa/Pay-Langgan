package payment

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func validSessionRequest() XenditSessionRequest {
	return XenditSessionRequest{
		ReferenceID:         "invoice_42",
		Amount:              150000,
		CustomerReferenceID: "customer_7",
		CustomerName:        "Alya Putri",
		CustomerEmail:       "alya@example.com",
		CustomerMobile:      "+628123456789",
		Description:         "Paket Pro Bulanan",
		SuccessReturnURL:    "https://paylanggan.example/checkout/success",
		CancelReturnURL:     "https://paylanggan.example/checkout/cancel",
	}
}

func TestXenditCreateCardSession(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/sessions" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("xnd_development_test:"))
		if r.Header.Get("Authorization") != wantAuth {
			t.Fatal("missing Xendit basic authentication")
		}
		var payload struct {
			ReferenceID string   `json:"reference_id"`
			SessionType string   `json:"session_type"`
			Mode        string   `json:"mode"`
			Amount      int64    `json:"amount"`
			Currency    string   `json:"currency"`
			Country     string   `json:"country"`
			Allowed     []string `json:"allowed_payment_channels"`
			Save        string   `json:"allow_save_payment_method"`
			Customer    struct {
				Email  string `json:"email"`
				Mobile string `json:"mobile_number"`
			} `json:"customer"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		if payload.ReferenceID != "invoice_42" || payload.SessionType != "PAY" || payload.Mode != "PAYMENT_LINK" || payload.Amount != 150000 || payload.Currency != "IDR" || payload.Country != "ID" || len(payload.Allowed) != 1 || payload.Allowed[0] != "CARDS" || payload.Save != "FORCED" || payload.Customer.Email == "" || payload.Customer.Mobile == "" {
			t.Fatalf("unexpected payload: %+v", payload)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"payment_session_id":"ps-123456789012345678901234","reference_id":"invoice_42","status":"ACTIVE","payment_link_url":"https://xen.to/test"}`))
	}))
	defer server.Close()
	provider, err := NewXenditProvider("xnd_development_test", "callback-secret", server.URL, true)
	if err != nil {
		t.Fatal(err)
	}
	provider.mock = false
	provider.httpClient = server.Client()
	response, err := provider.CreateCardSession(context.Background(), validSessionRequest())
	if err != nil {
		t.Fatal(err)
	}
	if response.Status != "ACTIVE" || response.PaymentLinkURL != "https://xen.to/test" {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func TestXenditSessionIsCardsOnlyAndValidatesInput(t *testing.T) {
	provider, err := NewXenditProvider("", "callback-secret", "https://api.xendit.co", true)
	if err != nil {
		t.Fatal(err)
	}
	request := validSessionRequest()
	response, err := provider.CreateCardSession(context.Background(), request)
	if err != nil || response.PaymentSessionID == "" || response.ReferenceID != request.ReferenceID {
		t.Fatalf("mock response=%+v error=%v", response, err)
	}

	tests := []struct {
		name   string
		mutate func(*XenditSessionRequest)
	}{
		{"missing reference", func(r *XenditSessionRequest) { r.ReferenceID = "" }},
		{"amount below card minimum", func(r *XenditSessionRequest) { r.Amount = 4999 }},
		{"amount above card maximum", func(r *XenditSessionRequest) { r.Amount = 200000001 }},
		{"invalid email", func(r *XenditSessionRequest) { r.CustomerEmail = "invalid" }},
		{"local phone", func(r *XenditSessionRequest) { r.CustomerMobile = "08123456789" }},
		{"insecure success URL", func(r *XenditSessionRequest) { r.SuccessReturnURL = "http://example.com/success" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := validSessionRequest()
			test.mutate(&request)
			if _, err := provider.CreateCardSession(context.Background(), request); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestXenditAPIErrorAndResponseValidation(t *testing.T) {
	for _, test := range []struct {
		name     string
		status   int
		body     string
		apiError bool
	}{
		{"provider error", http.StatusBadRequest, `{"error_code":"API_VALIDATION_ERROR","message":"invalid request"}`, true},
		{"malformed success", http.StatusCreated, `{`, false},
		{"incomplete success", http.StatusCreated, `{"payment_session_id":"ps-test","reference_id":"wrong"}`, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(test.status)
				_, _ = w.Write([]byte(test.body))
			}))
			defer server.Close()
			provider, _ := NewXenditProvider("key", "token", server.URL, true)
			provider.mock = false
			provider.httpClient = server.Client()
			_, err := provider.CreateCardSession(context.Background(), validSessionRequest())
			if err == nil {
				t.Fatal("expected error")
			}
			var apiError *XenditAPIError
			if errors.As(err, &apiError) != test.apiError {
				t.Fatalf("unexpected error type: %v", err)
			}
		})
	}
}

func TestXenditCallbackToken(t *testing.T) {
	provider, _ := NewXenditProvider("", "correct-token", "https://api.xendit.co", true)
	if !provider.VerifyCallbackToken("correct-token") || provider.VerifyCallbackToken("wrong-token") || provider.VerifyCallbackToken("") {
		t.Fatal("callback token verification failed")
	}
	withoutToken, _ := NewXenditProvider("", "", "https://api.xendit.co", true)
	if withoutToken.VerifyCallbackToken("") {
		t.Fatal("empty callback token must never verify")
	}
}

func TestXenditLiveConfiguration(t *testing.T) {
	if _, err := NewXenditProvider("", "token", "https://api.xendit.co", false); err == nil {
		t.Fatal("live provider accepted empty key")
	}
	if _, err := NewXenditProvider("key", "token", "http://api.xendit.test", false); err == nil {
		t.Fatal("live provider accepted HTTP")
	}
	if _, err := NewXenditProvider("key", "", "https://api.xendit.co", false); err == nil {
		t.Fatal("live provider accepted empty callback token")
	}
}
