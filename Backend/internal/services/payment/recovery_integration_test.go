package payment_test

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"pay-langgan/internal/database"
	"pay-langgan/internal/models"
	repo "pay-langgan/internal/repositories/payment"
	service "pay-langgan/internal/services/payment"
	provider "pay-langgan/pkg/payment"
)

type recoveryProvider struct {
	mu        sync.Mutex
	session   provider.XenditSessionResponse
	createErr error
	getErr    error
	creates   int
	onCreate  func()
}

func (p *recoveryProvider) CreateCardSession(_ context.Context, r provider.XenditSessionRequest) (*provider.XenditSessionResponse, error) {
	p.mu.Lock()
	p.creates++
	p.session = provider.XenditSessionResponse{PaymentSessionID: "ps-recovery", ReferenceID: r.ReferenceID, Amount: r.Amount, Currency: "IDR", Mode: "COMPONENTS", Status: "ACTIVE", ComponentsSDKKey: "sdk-not-persisted", ExpiresAt: time.Now().UTC().Add(30 * time.Minute)}
	session := p.session
	p.mu.Unlock()
	if p.onCreate != nil {
		p.onCreate()
	}
	return &session, p.createErr
}
func (p *recoveryProvider) GetCardSession(context.Context, string, string) (*provider.XenditSessionResponse, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	session := p.session
	return &session, p.getErr
}
func (p *recoveryProvider) VerifyCallbackToken(token string) bool { return token == "correct" }

func recoveryFixture(t *testing.T) (*database.DB, *service.CheckoutService, *recoveryProvider, *models.CheckoutLinkResponse) {
	t.Helper()
	db := paymentTestDB(t)
	p := &recoveryProvider{}
	s, err := service.NewCheckoutService(db, repo.NewCheckoutRepository(db), p, "http://localhost:5173", "https://paylanggan.example", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	link, err := s.CreateLink(context.Background(), 1, "test")
	if err != nil {
		t.Fatal(err)
	}
	return db, s, p, link
}

func startRecovery(s *service.CheckoutService, link *models.CheckoutLinkResponse) (*models.StartCheckoutResponse, error) {
	return s.Start(context.Background(), link.Token, models.StartCheckoutRequest{Name: "Alya", Email: "alya@example.com", Mobile: "+628123456789"})
}

func recoveryPayload(t *testing.T, p *recoveryProvider, status string) []byte {
	t.Helper()
	event := "payment_session.expired"
	if status == "COMPLETED" {
		event = "payment_session.completed"
	}
	data := p.session
	data.Status = status
	data.PaymentID = "py-recovery"
	data.PaymentTokenID = "pt-recovery"
	body, err := json.Marshal(map[string]any{"event": event, "data": data})
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func sqlExec(t *testing.T, db *database.DB, query string, args ...any) {
	t.Helper()
	if _, err := db.Exec(query, args...); err != nil {
		t.Fatal(err)
	}
}
func assertDBString(t *testing.T, db *database.DB, query, want string) {
	t.Helper()
	var got string
	if err := db.Get(&got, query); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("query %s: got %q want %q", query, got, want)
	}
}
func checkoutID(t *testing.T, db *database.DB) int64 {
	t.Helper()
	var id int64
	if err := db.Get(&id, `SELECT id FROM checkout_sessions WHERE invoice_id=1 ORDER BY id DESC LIMIT 1`); err != nil {
		t.Fatal(err)
	}
	return id
}

func TestRecoveryTimeoutAndWebhook(t *testing.T) {
	db, s, p, link := recoveryFixture(t)
	p.createErr = context.DeadlineExceeded
	if _, err := startRecovery(s, link); err == nil {
		t.Fatal("expected uncertain creation")
	}
	assertDBString(t, db, `SELECT status FROM checkout_sessions`, "creating")
	if _, err := startRecovery(s, link); err == nil {
		t.Fatal("uncertain session was recreated")
	}
	if _, err := s.CreateLink(context.Background(), 1, "test"); err == nil {
		t.Fatal("new checkout permitted during uncertainty")
	}
	if p.creates != 1 {
		t.Fatal("duplicate provider POST")
	}
	sqlExec(t, db, `UPDATE checkout_sessions SET updated_at=(NOW() AT TIME ZONE 'UTC')-INTERVAL '1 minute',next_reconcile_at=(NOW() AT TIME ZONE 'UTC')-INTERVAL '1 minute'`)
	result, err := s.RunMaintenance(context.Background(), 10)
	if err != nil || result.ReviewRequired != 1 {
		t.Fatalf("maintenance=%+v err=%v", result, err)
	}
	if _, err := s.ProcessWebhook(context.Background(), "correct", recoveryPayload(t, p, "COMPLETED")); err != nil {
		t.Fatal(err)
	}
	assertDBString(t, db, `SELECT status FROM invoices`, "paid")
	assertDBString(t, db, `SELECT status FROM checkout_sessions`, "completed")
	assertDBString(t, db, `SELECT count(*)::text FROM payments`, "1")
}

func TestRecoveryDefinitiveRejection(t *testing.T) {
	db, s, p, link := recoveryFixture(t)
	p.createErr = &provider.XenditAPIError{StatusCode: 400, Code: "API_VALIDATION_ERROR"}
	if _, err := startRecovery(s, link); err == nil {
		t.Fatal("rejection ignored")
	}
	assertDBString(t, db, `SELECT status FROM checkout_sessions`, "failed")
	if _, err := s.CreateLink(context.Background(), 1, "test"); err != nil {
		t.Fatal(err)
	}
}

func TestRecoveryManualSessionIdentity(t *testing.T) {
	db, s, p, link := recoveryFixture(t)
	p.createErr = errors.New("connection reset")
	_, _ = startRecovery(s, link)
	id := checkoutID(t, db)
	p.session.Amount++
	if err := s.RecoverCheckout(context.Background(), id, p.session.PaymentSessionID); err == nil {
		t.Fatal("wrong amount accepted")
	}
	assertDBString(t, db, `SELECT count(*)::text FROM payments`, "0")
	p.session.Amount--
	if err := s.RecoverCheckout(context.Background(), id, p.session.PaymentSessionID); err != nil {
		t.Fatal(err)
	}
	if _, err := startRecovery(s, link); err != nil {
		t.Fatal(err)
	}
	assertDBString(t, db, `SELECT status FROM checkout_sessions`, "active")
	assertDBString(t, db, `SELECT count(*)::text FROM payments`, "1")
	if p.creates != 1 {
		t.Fatal("recovery issued another provider POST")
	}
}

func TestRecoveryReconciliation(t *testing.T) {
	for _, status := range []string{"COMPLETED", "EXPIRED", "ACTIVE", "UNAVAILABLE"} {
		t.Run(status, func(t *testing.T) {
			db, s, p, link := recoveryFixture(t)
			if _, err := startRecovery(s, link); err != nil {
				t.Fatal(err)
			}
			sqlExec(t, db, `UPDATE checkout_sessions SET updated_at=(NOW() AT TIME ZONE 'UTC')-INTERVAL '1 minute',expires_at=(NOW() AT TIME ZONE 'UTC')-INTERVAL '1 minute'`)
			if status == "UNAVAILABLE" {
				p.getErr = errors.New("provider unavailable")
			} else {
				p.session.Status = status
			}
			p.session.PaymentID = "py-recovered"
			p.session.PaymentTokenID = "pt-recovered"
			result, err := s.RunMaintenance(context.Background(), 10)
			if err != nil || result.SessionsChecked != 1 {
				t.Fatalf("result=%+v err=%v", result, err)
			}
			switch status {
			case "COMPLETED":
				assertDBString(t, db, `SELECT status FROM invoices`, "paid")
				assertDBString(t, db, `SELECT token FROM payment_methods`, "pt-recovered")
			case "EXPIRED":
				assertDBString(t, db, `SELECT status FROM payments`, "expired")
				if _, err := s.CreateLink(context.Background(), 1, "test"); err != nil {
					t.Fatal(err)
				}
			default:
				assertDBString(t, db, `SELECT status FROM checkout_sessions`, "active")
				if _, err := s.CreateLink(context.Background(), 1, "test"); err == nil {
					t.Fatal("local expiry allowed duplicate checkout")
				}
			}
		})
	}
}

func TestRecoveryDurableWebhookRetry(t *testing.T) {
	db, s, p, link := recoveryFixture(t)
	if _, err := startRecovery(s, link); err != nil {
		t.Fatal(err)
	}
	sqlExec(t, db, `CREATE FUNCTION fail_payment_update() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'test transient failure'; END; $$;
 CREATE TRIGGER fail_payment_update BEFORE UPDATE ON payments FOR EACH ROW EXECUTE FUNCTION fail_payment_update()`)
	body := recoveryPayload(t, p, "COMPLETED")
	duplicate, err := s.ProcessWebhook(context.Background(), "correct", body)
	if err != nil || duplicate {
		t.Fatalf("receipt not acknowledged: %v", err)
	}
	assertDBString(t, db, `SELECT status FROM invoices`, "pending")
	assertDBString(t, db, `SELECT attempts::text FROM webhook_events`, "1")
	assertDBString(t, db, `SELECT processed::text FROM webhook_events`, "false")
	assertDBString(t, db, `SELECT (payload::text LIKE '%sdk-not-persisted%')::text FROM webhook_events`, "false")
	duplicate, err = s.ProcessWebhook(context.Background(), "correct", body)
	if err != nil || !duplicate {
		t.Fatal("duplicate not deduplicated")
	}
	sqlExec(t, db, `DROP TRIGGER fail_payment_update ON payments; UPDATE webhook_events SET next_attempt_at=(NOW() AT TIME ZONE 'UTC')-INTERVAL '1 minute'`)
	result, err := s.RunMaintenance(context.Background(), 10)
	if err != nil || result.WebhooksProcessed != 1 {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	assertDBString(t, db, `SELECT status FROM invoices`, "paid")
	assertDBString(t, db, `SELECT attempts::text FROM webhook_events`, "2")
	assertDBString(t, db, `SELECT count(*)::text FROM payment_methods`, "1")
}

func TestRecoveryWebhookRacesAndLateExpiration(t *testing.T) {
	db, s, p, link := recoveryFixture(t)
	p.onCreate = func() {
		if _, err := s.ProcessWebhook(context.Background(), "correct", recoveryPayload(t, p, "COMPLETED")); err != nil {
			t.Error(err)
		}
	}
	_, _ = startRecovery(s, link)
	assertDBString(t, db, `SELECT status FROM invoices`, "paid")
	assertDBString(t, db, `SELECT count(*)::text FROM payments`, "1")
	if _, err := s.ProcessWebhook(context.Background(), "correct", recoveryPayload(t, p, "EXPIRED")); err != nil {
		t.Fatal(err)
	}
	assertDBString(t, db, `SELECT status FROM checkout_sessions`, "completed")
	assertDBString(t, db, `SELECT status FROM payments`, "success")
}

func TestRecoveryConcurrentWebhookWorkers(t *testing.T) {
	db, s, p, link := recoveryFixture(t)
	if _, err := startRecovery(s, link); err != nil {
		t.Fatal(err)
	}
	body := recoveryPayload(t, p, "COMPLETED")
	var wg sync.WaitGroup
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := s.ProcessWebhook(context.Background(), "correct", body); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	assertDBString(t, db, `SELECT count(*)::text FROM webhook_events`, "1")
	assertDBString(t, db, `SELECT count(*)::text FROM payment_methods`, "1")
	assertDBString(t, db, `SELECT status FROM invoices`, "paid")
}

func TestRecoveryDeadLetterAndManualRetry(t *testing.T) {
	db, s, p, link := recoveryFixture(t)
	if _, err := startRecovery(s, link); err != nil {
		t.Fatal(err)
	}
	sqlExec(t, db, `CREATE FUNCTION fail_payment_update() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'test transient failure'; END; $$;
 CREATE TRIGGER fail_payment_update BEFORE UPDATE ON payments FOR EACH ROW EXECUTE FUNCTION fail_payment_update()`)
	if _, err := s.ProcessWebhook(context.Background(), "correct", recoveryPayload(t, p, "COMPLETED")); err != nil {
		t.Fatal(err)
	}
	sqlExec(t, db, `UPDATE webhook_events SET attempts=7,next_attempt_at=(NOW() AT TIME ZONE 'UTC')-INTERVAL '1 minute'`)
	_, _ = s.RunMaintenance(context.Background(), 10)
	assertDBString(t, db, `SELECT (dead_letter_at IS NOT NULL)::text FROM webhook_events`, "true")
	sqlExec(t, db, `DROP TRIGGER fail_payment_update ON payments`)
	var eventID int64
	if err := db.Get(&eventID, `SELECT id FROM webhook_events`); err != nil {
		t.Fatal(err)
	}
	if err := s.RetryWebhook(context.Background(), eventID); err != nil {
		t.Fatal(err)
	}
	assertDBString(t, db, `SELECT status FROM invoices`, "paid")
	if err := s.RetryWebhook(context.Background(), eventID); err == nil {
		t.Fatal("processed event replay permitted")
	}
}

func TestRecoveryAfterActivationFailure(t *testing.T) {
	db, s, p, link := recoveryFixture(t)
	sqlExec(t, db, `CREATE FUNCTION fail_payment_insert() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'activation failed'; END; $$;
 CREATE TRIGGER fail_payment_insert BEFORE INSERT ON payments FOR EACH ROW EXECUTE FUNCTION fail_payment_insert()`)
	if _, err := startRecovery(s, link); err == nil {
		t.Fatal("expected activation failure")
	}
	assertDBString(t, db, `SELECT provider_session_id FROM checkout_sessions`, "ps-recovery")
	assertDBString(t, db, `SELECT status FROM checkout_sessions`, "creating")
	sqlExec(t, db, `DROP TRIGGER fail_payment_insert ON payments;
 UPDATE checkout_sessions SET updated_at=(NOW() AT TIME ZONE 'UTC')-INTERVAL '1 minute'`)
	result, err := s.RunMaintenance(context.Background(), 10)
	if err != nil || result.SessionsChecked != 1 || result.Failures != 0 {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	assertDBString(t, db, `SELECT status FROM checkout_sessions`, "active")
	assertDBString(t, db, `SELECT count(*)::text FROM payments`, "1")
	if p.creates != 1 {
		t.Fatal("activation recovery sent another POST")
	}
}

func TestRecoveryRejectsMismatchedWebhook(t *testing.T) {
	db, s, p, link := recoveryFixture(t)
	if _, err := startRecovery(s, link); err != nil {
		t.Fatal(err)
	}
	p.session.Amount++
	body := recoveryPayload(t, p, "COMPLETED")
	if _, err := s.ProcessWebhook(context.Background(), "wrong", body); err == nil {
		t.Fatal("unauthenticated event accepted")
	}
	assertDBString(t, db, `SELECT count(*)::text FROM webhook_events`, "0")
	if _, err := s.ProcessWebhook(context.Background(), "correct", body); err != nil {
		t.Fatal(err)
	}
	assertDBString(t, db, `SELECT (dead_letter_at IS NOT NULL)::text FROM webhook_events`, "true")
	assertDBString(t, db, `SELECT status FROM invoices`, "pending")
	assertDBString(t, db, `SELECT status FROM payments`, "pending")
	// Independent provider reconciliation can settle the payment even when a
	// contradictory webhook is retained for investigation.
	p.session.Amount--
	p.session.Status = "COMPLETED"
	p.session.PaymentID = "py-verified"
	if err := s.RecoverCheckout(context.Background(), checkoutID(t, db), ""); err != nil {
		t.Fatal(err)
	}
	assertDBString(t, db, `SELECT status FROM invoices`, "paid")
}

func TestRecoveryConcurrentReconciliationAndWebhook(t *testing.T) {
	db, s, p, link := recoveryFixture(t)
	if _, err := startRecovery(s, link); err != nil {
		t.Fatal(err)
	}
	sqlExec(t, db, `UPDATE checkout_sessions SET updated_at=(NOW() AT TIME ZONE 'UTC')-INTERVAL '1 minute'`)
	p.session.Status = "COMPLETED"
	p.session.PaymentID = "py-recovery"
	p.session.PaymentTokenID = "pt-recovery"
	body := recoveryPayload(t, p, "COMPLETED")
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
	wg.Add(1)
	go func() {
		defer wg.Done()
		if _, err := s.ProcessWebhook(context.Background(), "correct", body); err != nil {
			t.Error(err)
		}
	}()
	wg.Wait()
	assertDBString(t, db, `SELECT status FROM invoices`, "paid")
	assertDBString(t, db, `SELECT count(*)::text FROM payments`, "1")
	assertDBString(t, db, `SELECT count(*)::text FROM payment_methods`, "1")
}
