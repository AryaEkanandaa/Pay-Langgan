package payment

import (
	"context"
	"testing"
	"time"

	provider "pay-langgan/pkg/payment"
)

type stubSessionProvider struct {
	token string
}

func (p *stubSessionProvider) CreateCardSession(context.Context, provider.XenditSessionRequest) (*provider.XenditSessionResponse, error) {
	return &provider.XenditSessionResponse{}, nil
}

func (p *stubSessionProvider) VerifyCallbackToken(token string) bool {
	return token == p.token
}

func (p *stubSessionProvider) GetCardSession(context.Context, string, string) (*provider.XenditSessionResponse, error) {
	return &provider.XenditSessionResponse{}, nil
}

func TestComponentsCheckoutResponse(t *testing.T) {
	expiry := time.Now().UTC().Add(30 * time.Minute)
	session := &provider.XenditSessionResponse{PaymentSessionID: "ps-test", Mode: "COMPONENTS", Status: "ACTIVE", ComponentsSDKKey: "sdk-test", ExpiresAt: expiry}
	response, err := componentsCheckoutResponse(session, expiry.Add(time.Hour))
	if err != nil || response.ComponentsSDKKey != "sdk-test" || !response.ExpiresAt.Equal(expiry) {
		t.Fatal("invalid components response")
	}
	for _, status := range []string{"COMPLETED", "EXPIRED", "CANCELED"} {
		session.Status = status
		if _, err := componentsCheckoutResponse(session, expiry); err == nil {
			t.Fatal("terminal session accepted")
		}
	}
	session.Status = "ACTIVE"
	if _, err := componentsCheckoutResponse(session, time.Now().Add(-time.Minute)); err == nil {
		t.Fatal("expired local checkout accepted")
	}
	session.ComponentsSDKKey = ""
	if _, err := componentsCheckoutResponse(session, expiry); err == nil {
		t.Fatal("missing SDK key accepted")
	}
}

func TestCheckoutTokenHash(t *testing.T) {
	token, err := secureToken(32)
	if err != nil {
		t.Fatal(err)
	}
	if len(token) < 40 || token == hashToken(token) || len(hashToken(token)) != 64 {
		t.Fatal("checkout token or hash has an unexpected format")
	}
	second, err := secureToken(32)
	if err != nil || second == token {
		t.Fatal("checkout tokens must be random")
	}
}

func TestWebhookRejectsUnauthenticatedAndInvalidEventsBeforeDatabase(t *testing.T) {
	service := &CheckoutService{provider: &stubSessionProvider{token: "correct"}}
	if _, err := service.ProcessWebhook(context.Background(), "wrong", []byte(`{}`)); err == nil {
		t.Fatal("invalid callback token was accepted")
	}
	if _, err := service.ProcessWebhook(context.Background(), "correct", []byte(`{`)); err == nil {
		t.Fatal("invalid JSON was accepted")
	}
	if _, err := service.ProcessWebhook(context.Background(), "correct", []byte(`{"event":"payment.capture"}`)); err == nil {
		t.Fatal("unsupported event was accepted")
	}
}

func TestCheckoutBaseURLValidation(t *testing.T) {
	for value, valid := range map[string]bool{
		"http://localhost:5173":            true,
		"https://paylanggan.example":       true,
		"javascript:alert(1)":              false,
		"https://example.com/path?token=x": false,
	} {
		if got := validBaseURL(value, false); got != valid {
			t.Fatalf("validBaseURL(%q)=%v", value, got)
		}
	}
	if validBaseURL("http://localhost:5173", true) {
		t.Fatal("Xendit return URL accepted HTTP")
	}
}

func TestCheckoutContactValidation(t *testing.T) {
	if err := validateCheckoutContact("Alya", "alya@example.com", "+628123456789"); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct{ name, email, mobile string }{
		{"", "alya@example.com", "+628123456789"},
		{"Alya", "invalid", "+628123456789"},
		{"Alya", "alya@example.com", "08123456789"},
	} {
		if err := validateCheckoutContact(test.name, test.email, test.mobile); err == nil {
			t.Fatalf("invalid contact accepted: %+v", test)
		}
	}
}
