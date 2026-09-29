package payment_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"pay-langgan/internal/database"
	handler "pay-langgan/internal/handlers/payment"
	billingrepo "pay-langgan/internal/repositories/billing"
	repo "pay-langgan/internal/repositories/payment"
	"pay-langgan/internal/routes"
	billingservice "pay-langgan/internal/services/billing"
	service "pay-langgan/internal/services/payment"
	"pay-langgan/internal/utils"
	provider "pay-langgan/pkg/payment"
)

type autopayProvider struct {
	*recoveryProvider
	chargeMu     sync.Mutex
	chargeCalls  int
	chargeErr    error
	chargeStatus string
	response     *provider.RecurringChargeResponse
}

func (p *autopayProvider) CreateRecurringCharge(_ context.Context, r provider.RecurringChargeRequest) (*provider.RecurringChargeResponse, error) {
	p.chargeMu.Lock()
	defer p.chargeMu.Unlock()
	p.chargeCalls++
	status := p.chargeStatus
	if status == "" {
		status = "SUCCEEDED"
	}
	p.response = &provider.RecurringChargeResponse{PaymentRequestID: "pr-recurring", ReferenceID: r.ReferenceID, Amount: r.Amount, Currency: "IDR", Status: status, PaymentID: "py-recurring", FailureCode: "CARD_DECLINED"}
	copy := *p.response
	return &copy, p.chargeErr
}
func (p *autopayProvider) GetRecurringCharge(context.Context, string) (*provider.RecurringChargeResponse, error) {
	p.chargeMu.Lock()
	defer p.chargeMu.Unlock()
	if p.response == nil {
		return nil, errors.New("missing request")
	}
	copy := *p.response
	return &copy, nil
}

func autopayFixture(t *testing.T) (*database.DB, *service.CheckoutService, *autopayProvider) {
	t.Helper()
	db := paymentTestDB(t)
	sqlExec(t, db, `INSERT INTO users(id,business_id,name,email,password,role) VALUES(1,'test','Finance','finance@example.com','not-a-real-password','finance');
 INSERT INTO payment_methods(id,customer_id,provider,token,is_default) VALUES(1,1,'xendit','pt-secret-token',TRUE);
 UPDATE subscriptions SET billing_cycle='monthly',billing_anchor=(NOW() AT TIME ZONE 'UTC')-INTERVAL '1 month';
 UPDATE invoices SET period_start=(NOW() AT TIME ZONE 'UTC')-INTERVAL '30 minutes',period_end=(NOW() AT TIME ZONE 'UTC')+INTERVAL '1 month',due_date=(NOW() AT TIME ZONE 'UTC')-INTERVAL '30 minutes'`)
	p := &autopayProvider{recoveryProvider: &recoveryProvider{}}
	s, err := service.NewCheckoutService(db, repo.NewCheckoutRepository(db), p, "http://localhost:5173", "https://paylanggan.example", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.SetAutopay(context.Background(), 1, "test", 1, service.AutopayRequest{Enabled: true, PaymentMethodID: 1, ConsentReference: "consent-document-123", ConsentedAt: time.Now().Add(-time.Hour), ConsentExpiresAt: time.Now().AddDate(1, 0, 0), MaxAmount: 200000})
	if err != nil {
		t.Fatal(err)
	}
	// Simulate a grant recorded before this invoice's billing period started.
	sqlExec(t, db, `UPDATE subscription_autopay SET effective_at=(NOW() AT TIME ZONE 'UTC')-INTERVAL '1 hour'`)
	s.EnableAutopay(true)
	return db, s, p
}

func TestAutopaySuccessAndNoDoubleCharge(t *testing.T) {
	db, s, p := autopayFixture(t)
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := s.RunMaintenance(context.Background(), 10); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	if p.chargeCalls != 1 {
		t.Fatalf("provider charge calls=%d", p.chargeCalls)
	}
	assertDBString(t, db, `SELECT status FROM invoices`, "paid")
	assertDBString(t, db, `SELECT status FROM payment_attempts`, "success")
	assertDBString(t, db, `SELECT count(*)::text FROM payments`, "1")
	items, total, err := s.PaymentHistory(context.Background(), "test", 1, 0, 1, 10, "")
	if err != nil || total != 1 || items[0].Status != "success" || items[0].PaidAt == nil {
		t.Fatalf("history=%+v err=%v", items, err)
	}
	body, _ := json.Marshal(items)
	if strings.Contains(string(body), "pt-secret-token") {
		t.Fatal("token leaked")
	}
}

func TestAutopayEligibility(t *testing.T) {
	for _, scenario := range []string{"disabled", "revoked", "paused", "cancelled", "expired consent", "old invoice", "future invoice", "manual checkout", "above limit"} {
		t.Run(scenario, func(t *testing.T) {
			db, s, p := autopayFixture(t)
			switch scenario {
			case "disabled":
				s.EnableAutopay(false)
			case "revoked":
				if _, err := s.SetAutopay(context.Background(), 1, "test", 1, service.AutopayRequest{Enabled: false}); err != nil {
					t.Fatal(err)
				}
			case "paused", "cancelled":
				sqlExec(t, db, `UPDATE subscriptions SET status=$1`, scenario)
			case "expired consent":
				sqlExec(t, db, `UPDATE subscription_autopay SET consented_at=(NOW() AT TIME ZONE 'UTC')-INTERVAL '2 days',consent_expires_at=(NOW() AT TIME ZONE 'UTC')-INTERVAL '1 day'`)
			case "old invoice":
				sqlExec(t, db, `UPDATE subscription_autopay SET effective_at=(NOW() AT TIME ZONE 'UTC')`)
			case "future invoice":
				sqlExec(t, db, `UPDATE invoices SET due_date=(NOW() AT TIME ZONE 'UTC')+INTERVAL '1 day'`)
			case "above limit":
				sqlExec(t, db, `UPDATE subscription_autopay SET max_amount=100000`)
			case "manual checkout":
				link, err := s.CreateLink(context.Background(), 1, "test")
				if err != nil {
					t.Fatal(err)
				}
				if _, err := startRecovery(s, link); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := s.RunMaintenance(context.Background(), 10); err != nil {
				t.Fatal(err)
			}
			if p.chargeCalls != 0 {
				t.Fatal("unapproved charge sent")
			}
			assertDBString(t, db, `SELECT status FROM invoices`, "pending")
			if scenario == "above limit" {
				assertDBString(t, db, `SELECT failure_code FROM autopay_charges`, "amount_outside_consent")
			}
		})
	}
}

func TestAutopayTimeoutRecoveryAndManualBlock(t *testing.T) {
	db, s, p := autopayFixture(t)
	p.chargeErr = context.DeadlineExceeded
	if _, err := s.RunMaintenance(context.Background(), 10); err != nil {
		t.Fatal(err)
	}
	assertDBString(t, db, `SELECT status FROM autopay_charges`, "review_required")
	if _, err := s.CreateLink(context.Background(), 1, "test"); err == nil {
		t.Fatal("uncertain charge permits checkout")
	}
	billing := billingservice.NewInvoiceService(db, billingrepo.NewInvoiceRepository(db))
	if _, err := billing.MarkPaid(1, "test"); err == nil {
		t.Fatal("uncertain charge permits mark-paid")
	}
	if _, err := s.RunMaintenance(context.Background(), 10); err != nil {
		t.Fatal(err)
	}
	if p.chargeCalls != 1 {
		t.Fatal("uncertain debit resent")
	}
	var id int64
	if err := db.Get(&id, `SELECT id FROM autopay_charges`); err != nil {
		t.Fatal(err)
	}
	if err := s.RecoverAutopay(context.Background(), id, "pr-recurring"); err != nil {
		t.Fatal(err)
	}
	assertDBString(t, db, `SELECT status FROM invoices`, "paid")
}

func TestAutopayDeclineAndReconciliation(t *testing.T) {
	for _, status := range []string{"FAILED", "ACCEPTING_PAYMENTS", "REQUIRES_ACTION"} {
		t.Run(status, func(t *testing.T) {
			db, s, p := autopayFixture(t)
			p.chargeStatus = status
			if _, err := s.RunMaintenance(context.Background(), 10); err != nil {
				t.Fatal(err)
			}
			assertDBString(t, db, `SELECT status FROM invoices`, "pending")
			if status == "FAILED" {
				if _, err := s.CreateLink(context.Background(), 1, "test"); err != nil {
					t.Fatal(err)
				}
				if _, err := s.RunMaintenance(context.Background(), 10); err != nil {
					t.Fatal(err)
				}
				if p.chargeCalls != 1 {
					t.Fatal("declined debit retried without authorization")
				}
			} else {
				p.response.Status = "SUCCEEDED"
				s.EnableAutopay(false) // Turning off new debits must not stop settlement.
				sqlExec(t, db, `UPDATE autopay_charges SET updated_at=(NOW() AT TIME ZONE 'UTC')-INTERVAL '1 minute',next_check_at=(NOW() AT TIME ZONE 'UTC')-INTERVAL '1 minute'`)
				if _, err := s.RunMaintenance(context.Background(), 10); err != nil {
					t.Fatal(err)
				}
				assertDBString(t, db, `SELECT status FROM invoices`, "paid")
			}
		})
	}
}

func TestPaymentControlsTenantAndRoles(t *testing.T) {
	_, s, _ := autopayFixture(t)
	if _, err := s.SetAutopay(context.Background(), 1, "other-business", 1, service.AutopayRequest{Enabled: false}); !errors.Is(err, utils.ErrNotFound) {
		t.Fatal("cross-tenant setting allowed", err)
	}
	items, total, err := s.PaymentHistory(context.Background(), "other-business", 0, 0, 1, 10, "")
	if err != nil || total != 0 || len(items) != 0 {
		t.Fatal("cross-tenant history leaked")
	}
	methods, err := s.PaymentMethods(context.Background(), "test", 1)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(methods)
	if strings.Contains(string(body), "token") {
		t.Fatal("method token leaked")
	}
	_, err = s.SetAutopay(context.Background(), 1, "test", 1, service.AutopayRequest{Enabled: true})
	if !errors.Is(err, utils.ErrBadRequest) {
		t.Fatal("consent-free activation allowed")
	}
	e := echo.New()
	routes.RegisterRoutes(e, &routes.AllHandlers{Checkout: handler.NewCheckoutHandler(s)}, "test-secret")
	for role, want := range map[string]int{"admin": 200, "finance": 200, "sales": 403, "super_admin": 403} {
		token, err := utils.GenerateToken("test-secret", time.Hour, 1, "test", "finance@example.com", role)
		if err != nil {
			t.Fatal(err)
		}
		r := httptest.NewRequest(http.MethodGet, "/api/v1/payments", nil)
		r.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		e.ServeHTTP(w, r)
		if w.Code != want {
			t.Fatalf("role=%s code=%d body=%s", role, w.Code, w.Body.String())
		}
	}
}

func TestCheckoutRateLimitSharedStorage(t *testing.T) {
	db, s, _, _ := recoveryFixture(t)
	var allowed atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ok, err := s.AllowCheckoutRequest(context.Background(), "session:token:secret", 5)
			if err != nil {
				t.Error(err)
			}
			if ok {
				allowed.Add(1)
			}
		}()
	}
	wg.Wait()
	if allowed.Load() != 5 {
		t.Fatalf("accepted %d requests", allowed.Load())
	}
	assertDBString(t, db, `SELECT (bucket_key LIKE '%secret%')::text FROM checkout_rate_limits`, "false")
	sqlExec(t, db, `UPDATE checkout_rate_limits SET window_start=(NOW() AT TIME ZONE 'UTC')-INTERVAL '2 minutes'`)
	if ok, err := s.AllowCheckoutRequest(context.Background(), "session:token:secret", 5); err != nil || !ok {
		t.Fatal("window did not reset", err)
	}
}
