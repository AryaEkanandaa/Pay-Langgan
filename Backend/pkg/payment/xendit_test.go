package payment

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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
		ComponentsOrigins:   []string{"https://paylanggan.example"},
		ReturnURL:           "https://paylanggan.example/checkout/done",
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
			ReferenceID      string                  `json:"reference_id"`
			SessionType      string                  `json:"session_type"`
			Mode             string                  `json:"mode"`
			Amount           int64                   `json:"amount"`
			Currency         string                  `json:"currency"`
			Country          string                  `json:"country"`
			Allowed          []string                `json:"allowed_payment_channels"`
			Save             string                  `json:"allow_save_payment_method"`
			Components       ComponentsConfiguration `json:"components_configuration"`
			SuccessReturnURL string                  `json:"success_return_url"`
			Customer         struct {
				Email  string `json:"email"`
				Mobile string `json:"mobile_number"`
			} `json:"customer"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		if payload.ReferenceID != "invoice_42" || payload.SessionType != "PAY" || payload.Mode != "COMPONENTS" || payload.Amount != 150000 || payload.Currency != "IDR" || payload.Country != "ID" || len(payload.Allowed) != 1 || payload.Allowed[0] != "CARDS" || payload.Save != "FORCED" || payload.Customer.Email == "" || payload.Customer.Mobile == "" {
			t.Fatalf("unexpected payload: %+v", payload)
		}
		if len(payload.Components.Origins) != 1 || payload.Components.Origins[0] != "https://paylanggan.example" || payload.Components.ReturnURL != validSessionRequest().ReturnURL || payload.SuccessReturnURL != "" {
			t.Fatal("invalid components configuration")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(XenditSessionResponse{PaymentSessionID: "ps-123456789012345678901234", ReferenceID: "invoice_42", Status: "ACTIVE", Mode: "COMPONENTS", ComponentsSDKKey: "sdk-test", ExpiresAt: time.Now().Add(time.Hour)})
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
	if response.Status != "ACTIVE" || response.ComponentsSDKKey != "sdk-test" || response.PaymentLinkURL != "" {
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
		{"insecure return URL", func(r *XenditSessionRequest) { r.ReturnURL = "http://example.com/success" }},
		{"missing origins", func(r *XenditSessionRequest) { r.ComponentsOrigins = nil }},
		{"wildcard origin", func(r *XenditSessionRequest) { r.ComponentsOrigins = []string{"https://*.example.com"} }},
		{"origin with path", func(r *XenditSessionRequest) { r.ComponentsOrigins = []string{"https://example.com/checkout"} }},
		{"remote HTTP origin", func(r *XenditSessionRequest) { r.ComponentsOrigins = []string{"http://example.com"} }},
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
		{"hosted response", http.StatusCreated, `{"payment_session_id":"ps-test","reference_id":"invoice_42","mode":"PAYMENT_LINK","status":"ACTIVE","payment_link_url":"https://xen.to/test"}`, false},
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

func TestXenditGetCardSession(t *testing.T) {
	for _, reference := range []string{"invoice_42", "another-invoice"} {
		t.Run(reference, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				key, _, ok := r.BasicAuth()
				if r.Method != http.MethodGet || r.URL.Path != "/sessions/ps-test" || !ok || key != "key" {
					t.Error("invalid session retrieval request")
				}
				_ = json.NewEncoder(w).Encode(XenditSessionResponse{PaymentSessionID: "ps-test", ReferenceID: reference, Mode: "COMPONENTS", Status: "ACTIVE", ComponentsSDKKey: "sdk-test", ExpiresAt: time.Now().Add(time.Hour)})
			}))
			defer server.Close()
			p, _ := NewXenditProvider("key", "token", server.URL, true)
			p.mock = false
			p.httpClient = server.Client()
			session, err := p.GetCardSession(context.Background(), "ps-test", "invoice_42")
			if reference != "invoice_42" {
				if err == nil {
					t.Fatal("mismatched session accepted")
				}
			} else if err != nil || session.ComponentsSDKKey != "sdk-test" {
				t.Fatal("failed to retrieve SDK key", err)
			}
		})
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

func TestSessionCreationRejected(t *testing.T) {
	for _, status := range []int{400, 401, 403, 404, 422} {
		if !SessionCreationRejected(&XenditAPIError{StatusCode: status}) {
			t.Fatalf("definitive HTTP %d not recognized", status)
		}
	}
	for _, status := range []int{408, 409, 429, 500, 502, 503} {
		if SessionCreationRejected(&XenditAPIError{StatusCode: status}) {
			t.Fatalf("uncertain HTTP %d treated as rejection", status)
		}
	}
	if SessionCreationRejected(context.DeadlineExceeded) || SessionCreationRejected(errors.New("invalid response")) {
		t.Fatal("ambiguous failure permits another session")
	}
	if !SessionCreationRejected(&ValidationError{Err: errors.New("invalid input")}) {
		t.Fatal("local validation not recognized")
	}
}
