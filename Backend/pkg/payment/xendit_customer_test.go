package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestXenditRepeatCustomerSession(t *testing.T) {
	for _, tc := range []struct {
		name   string
		lookup string
		wantOK bool
	}{
		{"reuse account customer", `{"data":[{"id":"cust-existing","reference_id":"customer_7"}]}`, true},
		{"other account not reusable", `{"data":[]}`, false},
		{"mismatched reference rejected", `{"data":[{"id":"cust-other","reference_id":"customer_8"}]}`, false},
		{"ambiguous lookup rejected", `{"data":[{"id":"cust-a","reference_id":"customer_7"},{"id":"cust-b","reference_id":"customer_7"}]}`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			posts := 0
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				user, _, ok := r.BasicAuth()
				if !ok || user != "tenant-key" {
					t.Error("lookup must use business key")
				}
				w.Header().Set("Content-Type", "application/json")
				if r.Method == http.MethodGet {
					if r.URL.Path != "/customers" || r.URL.Query().Get("reference_id") != "customer_7" || r.Header.Get("api-version") != "2020-10-31" {
						t.Error("invalid customer lookup")
					}
					fmt.Fprint(w, tc.lookup)
					return
				}
				posts++
				if posts == 1 {
					w.WriteHeader(409)
					fmt.Fprint(w, `{"error_code":"DUPLICATE_ERROR","message":"customer: reference_id already used"}`)
					return
				}
				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatal(err)
				}
				if body["customer_id"] != "cust-existing" || body["customer"] != nil || body["reference_id"] != "invoice_42" {
					t.Errorf("unsafe retry payload: %+v", body)
				}
				json.NewEncoder(w).Encode(XenditSessionResponse{PaymentSessionID: "ps-existing", ReferenceID: "invoice_42", Mode: "COMPONENTS", Status: "ACTIVE", ComponentsSDKKey: "sdk-test", ExpiresAt: time.Now().Add(time.Hour)})
			}))
			defer server.Close()
			provider, err := NewXenditProvider("tenant-key", "callback", server.URL, true)
			if err != nil {
				t.Fatal(err)
			}
			provider.mock = false
			provider.httpClient = server.Client()
			_, err = provider.CreateCardSession(context.Background(), validSessionRequest())
			if tc.wantOK {
				if err != nil || posts != 2 {
					t.Fatalf("reuse: posts=%d err=%v", posts, err)
				}
			} else if err == nil || posts != 1 || !SessionCreationRejected(err) {
				t.Fatalf("must reject without session retry: posts=%d err=%v", posts, err)
			}
		})
	}
}

func TestOnlyCustomerDuplicateIsDefinitivelyRejected(t *testing.T) {
	if SessionCreationRejected(&XenditAPIError{StatusCode: 409, Code: "DUPLICATE_ERROR", Message: "payment reference duplicate"}) {
		t.Fatal("payment conflict must remain uncertain")
	}
	if !SessionCreationRejected(&XenditAPIError{StatusCode: 409, Code: "DUPLICATE_ERROR", Message: "customer: reference_id duplicate"}) {
		t.Fatal("customer rejection is not an uncertain charge")
	}
}
