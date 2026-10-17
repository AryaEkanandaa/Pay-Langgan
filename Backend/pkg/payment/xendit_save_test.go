package payment

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestXenditSaveDoesNotCharge(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if string(body["session_type"]) != `"SAVE"` || string(body["amount"]) != "0" {
			t.Fatalf("not a zero-charge SAVE session: %s %s", body["session_type"], body["amount"])
		}
		if _, ok := body["capture_method"]; ok {
			t.Fatal("SAVE must not capture payments")
		}
		if _, ok := body["allow_save_payment_method"]; ok {
			t.Fatal("PAY-only flag sent")
		}
		_ = json.NewEncoder(w).Encode(XenditSessionResponse{SessionType: "SAVE", PaymentSessionID: "ps-save", ReferenceID: "invoice_42", Mode: "COMPONENTS", Status: "ACTIVE", ComponentsSDKKey: "sdk-test", ExpiresAt: time.Now().Add(time.Hour)})
	}))
	defer server.Close()
	p, _ := NewXenditProvider("test", "callback", server.URL, false)
	p.httpClient = server.Client()
	req := validSessionRequest()
	req.SessionType = "SAVE"
	req.Amount = 0
	if _, err := p.CreateCardSession(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	req.Amount = 5000
	if _, err := p.CreateCardSession(context.Background(), req); err == nil {
		t.Fatal("charged SAVE accepted")
	}
	req.SessionType = "PAY"
	req.Amount = 0
	if _, err := p.CreateCardSession(context.Background(), req); err == nil {
		t.Fatal("zero PAY accepted")
	}
}
