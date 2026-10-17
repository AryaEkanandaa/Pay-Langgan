package payment

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"pay-langgan/internal/models"
	billing "pay-langgan/internal/services/billing"
	"pay-langgan/internal/utils"
	provider "pay-langgan/pkg/payment"
)

// Enrollment is separate from invoice payment. In particular, a browser SDK
// registration receipt must never mark an invoice paid or enable automatic debit.
type trialCardRecord struct {
	ID             int64           `db:"id"`
	SubscriptionID int             `db:"subscription_id"`
	CustomerID     int             `db:"customer_id"`
	BusinessID     string          `db:"business_id"`
	Token          string          `db:"public_token"`
	Reference      string          `db:"reference_id"`
	Provider       string          `db:"provider"`
	Fingerprint    string          `db:"account_fingerprint"`
	Status         string          `db:"status"`
	SessionID      string          `db:"session_id"`
	Days           int             `db:"trial_days"`
	Amount         float64         `db:"amount"`
	Items          json.RawMessage `db:"items"`
	Expires        time.Time       `db:"expires_at"`
	EndsAt         *time.Time      `db:"trial_ends_at"`
	Name           string          `db:"customer_name"`
	Email          string          `db:"email"`
	Mobile         string          `db:"mobile"`
	Business       string          `db:"business_name"`
	Plan           string          `db:"plan_name"`
	Cycle          string          `db:"billing_cycle"`
}

const trialCardSelect = `SELECT t.id,t.subscription_id,s.customer_id,c.business_id,t.public_token,t.reference_id,t.provider,t.account_fingerprint,t.status,COALESCE(t.provider_session_id,'') session_id,t.trial_days,t.amount,t.items,t.expires_at,s.trial_ends_at,c.name customer_name,COALESCE(c.email,'') email,COALESCE(c.contact,'') mobile,b.name business_name,p.name plan_name,s.billing_cycle FROM trial_card_sessions t JOIN subscriptions s ON s.id=t.subscription_id JOIN customers c ON c.id=s.customer_id JOIN businesses b ON b.id=c.business_id JOIN plans p ON p.id=s.plan_id `

func trialAccount(g *BusinessGatewayRuntime) (string, string) {
	if g == nil {
		return "xendit", "test-provider"
	}
	// Bind enrollment to the exact credential/environment. Changing accounts or
	// rotating credentials requires starting a new enrollment, never token reuse.
	key := g.Config.XenditSecretKey
	if g.Config.Provider == "midtrans" {
		key = g.Config.MidtransServerKey + ":" + g.Config.MidtransClientKey
	}
	return g.Config.Provider, merchantKeyHash(g.BusinessID + ":" + g.Config.Provider + ":" + g.Config.Environment + ":" + key)
}

func (s *CheckoutService) prepareTrialCard(ctx context.Context, tx *sqlx.Tx, subscriptionID int, businessID string, days int, amount float64, items []models.SubscriptionPreviewItem) error {
	gateway, err := s.gatewayForBusiness(ctx, businessID)
	if err != nil {
		return err
	}
	name, fingerprint := trialAccount(gateway)
	token, err := secureToken(32)
	if err != nil {
		return err
	}
	ref, err := secureToken(20)
	if err != nil {
		return err
	}
	raw, err := json.Marshal(items)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO trial_card_sessions(subscription_id,public_token,reference_id,provider,account_fingerprint,trial_days,amount,items,expires_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`, subscriptionID, "trial_"+token, "trial_card_"+ref, name, fingerprint, days, amount, string(raw), time.Now().UTC().Add(s.ttl))
	return err
}

func (s *CheckoutService) trialRecord(ctx context.Context, token string) (*trialCardRecord, error) {
	if len(token) != 49 || !strings.HasPrefix(token, "trial_") {
		return nil, utils.ErrNotFound
	}
	var r trialCardRecord
	err := s.db.GetContext(ctx, &r, trialCardSelect+` WHERE t.public_token=$1`, token)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, utils.ErrNotFound
	}
	return &r, err
}

func (s *CheckoutService) checkTrialAccount(ctx context.Context, r *trialCardRecord) (*BusinessGatewayRuntime, error) {
	g, err := s.gatewayForBusiness(ctx, r.BusinessID)
	if err != nil {
		return nil, err
	}
	name, fingerprint := trialAccount(g)
	if name != r.Provider || fingerprint != r.Fingerprint {
		return nil, fmt.Errorf("%w: akun gateway berubah; mulai checkout baru setelah sesi ini berakhir", utils.ErrConflict)
	}
	return g, nil
}

func (s *CheckoutService) getTrialCard(ctx context.Context, token string) (*models.PublicCheckoutResponse, error) {
	r, err := s.trialRecord(ctx, token)
	if err != nil {
		return nil, err
	}
	if r.Provider == "xendit" && r.Status == "active" && r.SessionID != "" {
		// Bound provider polling across all open tabs. A webhook also reconciles.
		result, e := s.db.ExecContext(ctx, `UPDATE trial_card_sessions SET last_checked_at=NOW() WHERE id=$1 AND (last_checked_at IS NULL OR last_checked_at<NOW()-INTERVAL '5 seconds')`, r.ID)
		if e == nil {
			if n, _ := result.RowsAffected(); n == 1 {
				_ = s.reconcileTrialCard(ctx, r, r.SessionID)
			}
		}
		r, err = s.trialRecord(ctx, token)
		if err != nil {
			return nil, err
		}
	}
	if r.Status != "completed" && r.Status != "cancelled" && !r.Expires.After(time.Now()) {
		r.Status = "expired"
	}
	var snapshot []models.SubscriptionPreviewItem
	if err = json.Unmarshal(r.Items, &snapshot); err != nil {
		return nil, err
	}
	items := make([]models.CheckoutLineItem, 0, len(snapshot))
	for _, item := range snapshot {
		items = append(items, models.CheckoutLineItem{Type: item.Type, Description: item.Name, Quantity: item.Quantity, Subtotal: item.Subtotal})
	}
	return &models.PublicCheckoutResponse{CardRegistration: true, PaymentProvider: r.Provider, BillingCycle: r.Cycle, TrialDays: r.Days, TrialEndsAt: r.EndsAt, Items: items, Amount: r.Amount, InvoiceStatus: "not_issued", CheckoutStatus: r.Status, CustomerName: r.Name, CustomerEmail: &r.Email, CustomerMobile: &r.Mobile, BusinessName: r.Business, PlanName: r.Plan, ExpiresAt: r.Expires}, nil
}

func (s *CheckoutService) startTrialCard(ctx context.Context, token string, req models.StartCheckoutRequest) (*models.StartCheckoutResponse, error) {
	r, err := s.trialRecord(ctx, token)
	if err != nil {
		return nil, err
	}
	if !r.Expires.After(time.Now()) || r.Status == "completed" || r.Status == "expired" || r.Status == "cancelled" {
		return nil, fmt.Errorf("%w: sesi pendaftaran kartu telah berakhir", utils.ErrConflict)
	}
	if (r.Provider == "xendit" || r.Provider == "midtrans") && (r.Cycle != "monthly" && r.Cycle != "yearly" || r.Amount < 5000 || r.Amount > 200000000 || r.Amount != float64(int64(r.Amount))) {
		return nil, fmt.Errorf("%w: nilai tagihan setelah trial belum memenuhi batas debit otomatis kartu", utils.ErrBadRequest)
	}
	gateway, err := s.checkTrialAccount(ctx, r)
	if err != nil {
		return nil, err
	}
	if r.Status == "created" && !req.CardConsent {
		return nil, fmt.Errorf("%w: persetujuan penyimpanan kartu diperlukan", utils.ErrBadRequest)
	}
	returnURL := s.returnBaseURL + "/checkout/" + token
	if req.ParentOrigin != "" {
		parent, e := url.Parse(req.ParentOrigin)
		if e != nil || parent.Host == "" || parent.User != nil || parent.String() != parent.Scheme+"://"+parent.Host || (parent.Scheme != "https" && !(parent.Scheme == "http" && (parent.Hostname() == "localhost" || parent.Hostname() == "127.0.0.1"))) {
			return nil, utils.ErrBadRequest
		}
		returnURL += "?parent_origin=" + url.QueryEscape(req.ParentOrigin)
	}
	if r.Provider == "midtrans" {
		if gateway == nil {
			return nil, utils.ErrConflict
		}
		result, e := s.db.ExecContext(ctx, `UPDATE trial_card_sessions SET status='active',consented_at=COALESCE(consented_at,NOW()) WHERE id=$1 AND status IN ('created','active') AND expires_at>NOW()`, r.ID)
		if e != nil {
			return nil, e
		}
		if n, _ := result.RowsAffected(); n != 1 {
			return nil, utils.ErrConflict
		}
		return &models.StartCheckoutResponse{CardRegistration: true, Provider: "midtrans", Mode: "CORE", ClientKey: gateway.Config.MidtransClientKey, Environment: gateway.Config.Environment, ExpiresAt: r.Expires, Status: "active"}, nil
	}
	if r.Status == "creating" {
		// Xendit's request has a 15-second client timeout. A creating row older
		// than two minutes cannot still have an in-flight request in this API
		// process. Rotate the provider reference before retrying because a timed
		// out SAVE request may have reached Xendit without returning its session ID.
		newReference, tokenErr := secureToken(20)
		if tokenErr != nil {
			return nil, tokenErr
		}
		newReference = "trial_card_" + newReference
		result, updateErr := s.db.ExecContext(ctx, `UPDATE trial_card_sessions SET status='created',reference_id=$2 WHERE id=$1 AND status='creating' AND provider_session_id IS NULL AND consented_at<=NOW()-INTERVAL '2 minutes' AND expires_at>NOW()`, r.ID, newReference)
		if updateErr != nil {
			return nil, updateErr
		}
		if rows, _ := result.RowsAffected(); rows == 1 {
			r.Status = "created"
			r.Reference = newReference
		} else {
			return nil, fmt.Errorf("%w: pendaftaran kartu masih diproses. Tunggu sebentar lalu coba lagi", utils.ErrConflict)
		}
	}
	account, err := s.xenditForBusiness(ctx, r.BusinessID)
	if err != nil {
		return nil, err
	}
	var session *provider.XenditSessionResponse
	if r.Status == "active" && r.SessionID != "" {
		session, err = account.GetCardSession(ctx, r.SessionID, r.Reference)
	} else {
		result, e := s.db.ExecContext(ctx, `UPDATE trial_card_sessions SET status='creating',consented_at=NOW() WHERE id=$1 AND status='created' AND expires_at>NOW()`, r.ID)
		if e != nil {
			return nil, e
		}
		if n, _ := result.RowsAffected(); n != 1 {
			return nil, fmt.Errorf("%w: pendaftaran sedang diperiksa; tunggu konfirmasi atau mulai ulang setelah sesi berakhir", utils.ErrConflict)
		}
		app, _ := url.Parse(s.publicBaseURL)
		frequency := 28
		if r.Cycle == "yearly" {
			frequency = 365
		}
		expiry := time.Now().UTC().AddDate(10, 0, 0).Format("2006-01-02")
		session, err = account.CreateCardSession(ctx, provider.XenditSessionRequest{SessionType: "SAVE", ReferenceID: r.Reference, Amount: 0, CustomerReferenceID: fmt.Sprintf("paylanggan_%s_customer_%d", r.BusinessID, r.CustomerID), CustomerName: r.Name, CustomerEmail: r.Email, CustomerMobile: r.Mobile, Description: "Pendaftaran kartu trial - " + r.Plan, ComponentsOrigins: []string{app.Scheme + "://" + app.Host}, ReturnURL: returnURL, RecurringFrequency: frequency, RecurringExpiry: expiry})
		if err != nil {
			// Never blindly recreate a session after a timeout: its webhook may arrive.
			if provider.SessionCreationRejected(err) {
				_, _ = s.db.ExecContext(ctx, `UPDATE trial_card_sessions SET status='created' WHERE id=$1 AND status='creating'`, r.ID)
			}
			return nil, fmt.Errorf("prepare trial card session: %w", err)
		}
		if session == nil || session.PaymentSessionID == "" {
			return nil, fmt.Errorf("missing trial card session")
		}
		_, err = s.db.ExecContext(ctx, `UPDATE trial_card_sessions SET status='active',provider_session_id=$2,expires_at=LEAST(expires_at,$3) WHERE id=$1 AND status='creating'`, r.ID, session.PaymentSessionID, session.ExpiresAt)
	}
	if err != nil {
		return nil, err
	}
	response, err := componentsCheckoutResponse(session, r.Expires)
	if response != nil {
		response.CardRegistration = true
	}
	return response, err
}

func validTrialSaveSession(r *trialCardRecord, session *provider.XenditSessionResponse) bool {
	return session != nil && session.SessionType == "SAVE" && session.ReferenceID == r.Reference && session.PaymentSessionID != "" && (r.SessionID == "" || session.PaymentSessionID == r.SessionID) && session.Amount == 0 && session.Currency == "IDR"
}

func (s *CheckoutService) reconcileTrialCard(ctx context.Context, r *trialCardRecord, sessionID string) error {
	if _, err := s.checkTrialAccount(ctx, r); err != nil {
		return err
	}
	account, err := s.xenditForBusiness(ctx, r.BusinessID)
	if err != nil {
		return err
	}
	session, err := account.GetCardSession(ctx, sessionID, r.Reference)
	if err != nil {
		return err
	}
	if !validTrialSaveSession(r, session) {
		return fmt.Errorf("%w: mismatched card registration session", utils.ErrBadRequest)
	}
	if session.Status == "COMPLETED" && session.PaymentTokenID != "" {
		return s.completeTrialCard(ctx, r, session.PaymentTokenID, "provider_confirmed", session.PaymentSessionID)
	}
	if session.Status == "EXPIRED" || session.Status == "CANCELED" {
		_, err = s.db.ExecContext(ctx, `UPDATE trial_card_sessions SET expires_at=NOW() WHERE id=$1 AND status!='completed'`, r.ID)
		if err == nil {
			err = s.releaseExpiredTrialCards(ctx, s.db, 100)
		}
	}
	return err
}

// Abandoned enrollment releases a reserved coupon exactly once. Expiring the
// pending subscription also keeps it out of the active customer count.
func (s *CheckoutService) releaseExpiredTrialCards(ctx context.Context, db sqlx.ExtContext, limit int) error {
	_, err := db.ExecContext(ctx, `WITH candidates AS (
 SELECT id,subscription_id FROM trial_card_sessions
 WHERE status IN ('created','creating','active') AND expires_at<=NOW()
 ORDER BY expires_at,id FOR UPDATE SKIP LOCKED LIMIT $1
), expired AS (
 UPDATE trial_card_sessions t SET status='expired' FROM candidates x
 WHERE t.id=x.id RETURNING t.subscription_id
), released AS (
 UPDATE coupons c SET used_count=GREATEST(0,c.used_count-q.uses)
 FROM (SELECT sc.coupon_id,count(*)::int uses FROM subscription_coupons sc JOIN expired e ON e.subscription_id=sc.subscription_id GROUP BY sc.coupon_id) q
 WHERE c.id=q.coupon_id RETURNING c.id
)
 UPDATE subscriptions s SET status='cancelled' FROM expired e
 WHERE s.id=e.subscription_id AND s.status='pending' AND s.trial_ends_at IS NULL`, limit)
	return err
}

// Recover a completed SAVE session when the shopper closed the page and the
// webhook was delayed. Provider state remains the source of truth.
func (s *CheckoutService) reconcileOpenTrialCards(ctx context.Context, limit int) int {
	failures := 0
	if limit > 5 {
		limit = 5
	}
	for i := 0; i < limit; i++ {
		var id int64
		err := s.db.QueryRowContext(ctx, `WITH candidate AS (
 SELECT id FROM trial_card_sessions WHERE provider='xendit' AND status='active'
 AND provider_session_id IS NOT NULL AND expires_at>NOW()
 AND (last_checked_at IS NULL OR last_checked_at<NOW()-INTERVAL '30 seconds')
 ORDER BY id FOR UPDATE SKIP LOCKED LIMIT 1)
 UPDATE trial_card_sessions t SET last_checked_at=NOW() FROM candidate c
 WHERE t.id=c.id RETURNING t.id`).Scan(&id)
		if errors.Is(err, sql.ErrNoRows) {
			break
		}
		if err != nil {
			failures++
			break
		}
		var record trialCardRecord
		if err = s.db.GetContext(ctx, &record, trialCardSelect+` WHERE t.id=$1`, id); err == nil {
			err = s.reconcileTrialCard(ctx, &record, record.SessionID)
		}
		if err != nil {
			failures++
		}
		if ctx.Err() != nil {
			break
		}
	}
	return failures
}

// Authenticated webhooks route here before invoice lookup. Re-fetching the SAVE
// session binds the result to this merchant, reference and zero-charge purpose.
func (s *CheckoutService) trialCardWebhook(ctx context.Context, businessID string, event xenditSessionWebhook) (bool, error) {
	var r trialCardRecord
	err := s.db.GetContext(ctx, &r, trialCardSelect+` WHERE t.reference_id=$1`, event.Data.ReferenceID)
	if err != nil {
		return false, utils.ErrNotFound
	}
	if (businessID != "" && r.BusinessID != businessID) || r.Provider != "xendit" {
		return false, utils.ErrUnauthorized
	}
	if r.Status == "completed" {
		return true, nil
	}
	return false, s.reconcileTrialCard(ctx, &r, event.Data.PaymentSessionID)
}

type MidtransTrialRegistration struct {
	SavedTokenID  string `json:"saved_token_id"`
	TransactionID string `json:"transaction_id"`
}

var midtransSavedToken = regexp.MustCompile(`^[A-Za-z0-9_-]{16,256}$`)
var midtransRegistrationID = regexp.MustCompile(`^[a-fA-F0-9-]{32,36}$`)

func (s *CheckoutService) RegisterMidtransTrialCard(ctx context.Context, token string, req MidtransTrialRegistration) (*models.PublicCheckoutResponse, error) {
	if !midtransSavedToken.MatchString(req.SavedTokenID) || !strings.ContainsAny(req.SavedTokenID, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_-") || !midtransRegistrationID.MatchString(req.TransactionID) {
		return nil, utils.ErrBadRequest
	}
	r, err := s.trialRecord(ctx, token)
	if err != nil {
		return nil, err
	}
	if r.Provider != "midtrans" {
		return nil, utils.ErrConflict
	}
	if _, err = s.checkTrialAccount(ctx, r); err != nil {
		return nil, err
	}
	// Register Card returns a reusable token via the browser SDK, not a signed
	// payment/3DS proof. Store the assurance as 'registered', never 'verified'.
	// Actual charging must use merchant-authenticated APIs and validate their result.
	if err = s.completeTrialCard(ctx, r, req.SavedTokenID, "registered", req.TransactionID); err != nil {
		return nil, err
	}
	return s.getTrialCard(ctx, token)
}

func (s *CheckoutService) completeTrialCard(ctx context.Context, r *trialCardRecord, savedToken, level, sessionID string) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var status string
	var consent *time.Time
	var expiry time.Time
	if err = tx.QueryRowContext(ctx, `SELECT status,consented_at,expires_at FROM trial_card_sessions WHERE id=$1 FOR UPDATE`, r.ID).Scan(&status, &consent, &expiry); err != nil {
		return err
	}
	if status == "completed" {
		return nil
	}
	if (status != "active" && status != "creating") || consent == nil || !expiry.After(time.Now()) {
		return utils.ErrConflict
	}
	if r.Provider != "xendit" && r.Provider != "midtrans" {
		return fmt.Errorf("unsupported trial card provider")
	}
	if r.Cycle != "monthly" && r.Cycle != "yearly" || r.Amount < 5000 || r.Amount > 200000000 || r.Amount != float64(int64(r.Amount)) {
		return fmt.Errorf("trial cannot be enrolled for automatic debit")
	}
	if r.Provider == "xendit" && level != "provider_confirmed" || r.Provider == "midtrans" && level != "registered" {
		return fmt.Errorf("trial card registration status is invalid for %s", r.Provider)
	}
	now := time.Now().UTC()
	end := now.AddDate(0, 0, r.Days)
	result, err := tx.ExecContext(ctx, `UPDATE subscriptions SET status='trial',start_date=$2,billing_anchor=$3,next_billing_date=$3,trial_ends_at=$3 WHERE id=$1 AND status='pending' AND trial_ends_at IS NULL`, r.SubscriptionID, now, end)
	if err != nil {
		return err
	}
	if n, _ := result.RowsAffected(); n != 1 {
		return utils.ErrConflict
	}
	if _, err = tx.ExecContext(ctx, `UPDATE trial_card_sessions SET status='completed',provider_session_id=$2,registration_token=$3,verification_level=$4,completed_at=$5 WHERE id=$1`, r.ID, sessionID, savedToken, level, now); err != nil {
		return err
	}
	var paymentMethodID int
	err = tx.QueryRowContext(ctx, `INSERT INTO payment_methods(customer_id,provider,token,is_default,created_at)
 VALUES($1,$2,$3,NOT EXISTS(SELECT 1 FROM payment_methods WHERE customer_id=$1 AND is_default),(NOW() AT TIME ZONE 'UTC'))
 ON CONFLICT(provider,token) DO UPDATE SET is_default=payment_methods.is_default
 WHERE payment_methods.customer_id=EXCLUDED.customer_id RETURNING id`, r.CustomerID, r.Provider, savedToken).Scan(&paymentMethodID)
	if err != nil {
		return fmt.Errorf("save %s trial payment token: %w", r.Provider, err)
	}
	consentExpiry := now.AddDate(10, 0, 0)
	consentReference := "trial-autopay-v1:" + r.Reference
	if _, err = tx.ExecContext(ctx, `INSERT INTO subscription_autopay(subscription_id,payment_method_id,enabled,consent_reference,consented_at,consent_expires_at,max_amount,updated_by,effective_at,updated_at)
 VALUES($1,$2,TRUE,$3,$4,$5,$6,NULL,(NOW() AT TIME ZONE 'UTC'),(NOW() AT TIME ZONE 'UTC'))
 ON CONFLICT(subscription_id) DO UPDATE SET payment_method_id=EXCLUDED.payment_method_id,enabled=TRUE,consent_reference=EXCLUDED.consent_reference,
 consented_at=EXCLUDED.consented_at,consent_expires_at=EXCLUDED.consent_expires_at,max_amount=EXCLUDED.max_amount,
	 updated_by=NULL,effective_at=(NOW() AT TIME ZONE 'UTC'),updated_at=(NOW() AT TIME ZONE 'UTC')`, r.SubscriptionID, paymentMethodID, consentReference, consent.UTC(), consentExpiry, int64(r.Amount)); err != nil {
		return fmt.Errorf("enable consented %s trial autopay: %w", r.Provider, err)
	}
	consentAudit, _ := json.Marshal(map[string]any{"consent_version": "trial-autopay-v1", "provider": r.Provider, "billing_cycle": r.Cycle, "amount_limit": int64(r.Amount), "consented_at": consent.UTC(), "expires_at": consentExpiry})
	if _, err = tx.ExecContext(ctx, `INSERT INTO audit_logs(business_id,action,entity_type,entity_id,new_value) VALUES($1,'customer_consented_autopay','subscription',$2,$3::jsonb)`, r.BusinessID, fmt.Sprint(r.SubscriptionID), string(consentAudit)); err != nil {
		return fmt.Errorf("record trial autopay consent: %w", err)
	}
	var items []models.SubscriptionPreviewItem
	if err = json.Unmarshal(r.Items, &items); err != nil {
		return err
	}
	if err = billing.QueueTrialConfirmation(ctx, tx, r.SubscriptionID, r.Amount, items); err != nil {
		return err
	}
	return tx.Commit()
}
