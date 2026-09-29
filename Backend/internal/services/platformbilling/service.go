package platformbilling

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"pay-langgan/internal/database"
	billing "pay-langgan/internal/services/billing"
	"pay-langgan/internal/utils"
	provider "pay-langgan/pkg/payment"
)

const MonthlyPrice int64 = 300000
const PlanName = "PayLanggan Bulanan"

type Provider interface {
	CreateHostedSession(context.Context, provider.HostedSessionRequest) (*provider.XenditSessionResponse, error)
	GetHostedSession(context.Context, string) (*provider.XenditSessionResponse, error)
	CreateRecurringCharge(context.Context, provider.RecurringChargeRequest) (*provider.RecurringChargeResponse, error)
	GetRecurringCharge(context.Context, string) (*provider.RecurringChargeResponse, error)
	VerifyCallbackToken(string) bool
}

type Service struct {
	db        *database.DB
	provider  Provider
	returnURL string
}

func NewService(db *database.DB, p Provider, publicURL, returnURL string) (*Service, error) {
	publicURL = strings.TrimRight(strings.TrimSpace(publicURL), "/")
	returnURL = strings.TrimRight(strings.TrimSpace(returnURL), "/")
	if db == nil || p == nil || !strings.HasPrefix(returnURL, "https://") {
		return nil, fmt.Errorf("platform billing requires a database, provider, and HTTPS return URL")
	}
	return &Service{db: db, provider: p, returnURL: returnURL}, nil
}

type Plan struct {
	Code         string `json:"code"`
	Name         string `json:"name"`
	Amount       int64  `json:"amount"`
	Currency     string `json:"currency"`
	BillingCycle string `json:"billing_cycle"`
}

func CurrentPlan() Plan { return Plan{"paylanggan_monthly", PlanName, MonthlyPrice, "IDR", "monthly"} }

type Invoice struct {
	ID          int64      `db:"id" json:"id"`
	BusinessID  string     `db:"business_id" json:"-"`
	Amount      int64      `db:"amount" json:"amount"`
	Status      string     `db:"status" json:"status"`
	PeriodStart time.Time  `db:"period_start" json:"period_start"`
	PeriodEnd   time.Time  `db:"period_end" json:"period_end"`
	DueDate     time.Time  `db:"due_date" json:"due_date"`
	PaidAt      *time.Time `db:"paid_at" json:"paid_at"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
}

type Subscription struct {
	BusinessID      string     `db:"business_id" json:"business_id"`
	Status          string     `db:"status" json:"status"`
	PaidThrough     *time.Time `db:"paid_through" json:"paid_through"`
	NextBillingDate *time.Time `db:"next_billing_date" json:"next_billing_date"`
	AutoRenew       bool       `json:"auto_renew"`
	ConsentAt       *time.Time `db:"auto_renew_consent_at" json:"consent_at"`
	PendingInvoice  *Invoice   `json:"pending_invoice,omitempty"`
	CheckoutURL     string     `json:"checkout_url,omitempty"`
}

func (s *Service) Status(ctx context.Context, businessID string) (*Subscription, error) {
	var sub Subscription
	err := s.db.GetContext(ctx, &sub, `SELECT business_id,status,paid_through,next_billing_date,auto_renew_consent_at
 FROM platform_subscriptions WHERE business_id=$1`, businessID)
	if err == sql.ErrNoRows {
		return &Subscription{BusinessID: businessID, Status: "not_subscribed"}, nil
	}
	if err != nil {
		return nil, err
	}
	var revoked *time.Time
	if err := s.db.GetContext(ctx, &revoked, `SELECT auto_renew_revoked_at FROM platform_subscriptions WHERE business_id=$1`, businessID); err != nil {
		return nil, err
	}
	var token *string
	if err := s.db.GetContext(ctx, &token, `SELECT payment_token_id FROM platform_subscriptions WHERE business_id=$1`, businessID); err != nil {
		return nil, err
	}
	sub.AutoRenew = sub.ConsentAt != nil && revoked == nil && token != nil && *token != ""
	var inv Invoice
	err = s.db.GetContext(ctx, &inv, `SELECT id,business_id,amount,status,period_start,period_end,due_date,paid_at,created_at
 FROM platform_invoices WHERE business_id=$1 AND status='pending' ORDER BY id DESC LIMIT 1`, businessID)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if err == nil {
		sub.PendingInvoice = &inv
		_ = s.db.GetContext(ctx, &sub.CheckoutURL, `SELECT payment_link_url FROM platform_checkout_sessions
 WHERE invoice_id=$1 AND status='active' AND payment_link_url IS NOT NULL ORDER BY id DESC LIMIT 1`, inv.ID)
	}
	return &sub, nil
}

func (s *Service) Invoices(ctx context.Context, businessID string) ([]Invoice, error) {
	items := []Invoice{}
	err := s.db.SelectContext(ctx, &items, `SELECT id,business_id,amount,status,period_start,period_end,due_date,paid_at,created_at
 FROM platform_invoices WHERE business_id=$1 ORDER BY id DESC LIMIT 100`, businessID)
	return items, err
}

type Checkout struct {
	InvoiceID   int64  `json:"invoice_id"`
	Amount      int64  `json:"amount"`
	CheckoutURL string `json:"checkout_url"`
	Status      string `json:"status"`
}

func (s *Service) Start(ctx context.Context, businessID string, userID int, consent bool) (*Checkout, error) {
	if businessID == "" || userID < 1 || !consent {
		return nil, fmt.Errorf("%w: explicit monthly auto-debit consent is required", utils.ErrBadRequest)
	}
	var payer struct {
		Name  string `db:"name"`
		Email string `db:"email"`
	}
	if err := s.db.GetContext(ctx, &payer, `SELECT name,email FROM users WHERE id=$1 AND business_id=$2 AND role='admin'`, userID, businessID); err != nil {
		return nil, fmt.Errorf("%w: business admin not found", utils.ErrUnauthorized)
	}
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `INSERT INTO platform_subscriptions(business_id) VALUES($1) ON CONFLICT DO NOTHING`, businessID); err != nil {
		return nil, err
	}
	var sub struct {
		Status      string     `db:"status"`
		PaidThrough *time.Time `db:"paid_through"`
	}
	if err = tx.GetContext(ctx, &sub, `SELECT status,paid_through FROM platform_subscriptions WHERE business_id=$1 FOR UPDATE`, businessID); err != nil {
		return nil, err
	}
	if sub.Status == "cancelled" {
		return nil, fmt.Errorf("%w: cancelled platform subscription requires operator review", utils.ErrConflict)
	}
	if sub.PaidThrough != nil && sub.PaidThrough.After(time.Now().UTC()) {
		return nil, fmt.Errorf("%w: platform subscription is already paid", utils.ErrConflict)
	}
	var invoice Invoice
	err = tx.GetContext(ctx, &invoice, `SELECT id,business_id,amount,status,period_start,period_end,due_date,paid_at,created_at
 FROM platform_invoices WHERE business_id=$1 AND status='pending' ORDER BY id DESC LIMIT 1 FOR UPDATE`, businessID)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if err == sql.ErrNoRows {
		start := time.Now().UTC()
		end, err := billing.NextPeriod(start, start, "monthly")
		if err != nil {
			return nil, err
		}
		if err = tx.GetContext(ctx, &invoice, `INSERT INTO platform_invoices(business_id,period_start,period_end,amount,due_date)
 VALUES($1,$2,$3,$4,$2) RETURNING id,business_id,amount,status,period_start,period_end,due_date,paid_at,created_at`,
			businessID, start, end, MonthlyPrice); err != nil {
			return nil, err
		}
	} else if invoice.PeriodEnd.Before(time.Now().UTC()) {
		return nil, fmt.Errorf("%w: old platform invoice requires operator review", utils.ErrConflict)
	}
	var chargeStatus string
	err = tx.GetContext(ctx, &chargeStatus, `SELECT status FROM platform_autopay_charges WHERE invoice_id=$1`, invoice.ID)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if err == nil && chargeStatus != "failed" {
		return nil, fmt.Errorf("%w: automatic charge is in progress or requires review", utils.ErrConflict)
	}
	var existing struct {
		Status string  `db:"status"`
		URL    *string `db:"payment_link_url"`
	}
	err = tx.GetContext(ctx, &existing, `SELECT status,payment_link_url FROM platform_checkout_sessions
 WHERE invoice_id=$1 AND status IN ('creating','active') ORDER BY id DESC LIMIT 1`, invoice.ID)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if err == nil {
		if existing.Status == "active" && existing.URL != nil && *existing.URL != "" {
			return &Checkout{invoice.ID, invoice.Amount, *existing.URL, "active"}, nil
		}
		return nil, fmt.Errorf("%w: checkout creation is in progress; reconcile before retrying", utils.ErrConflict)
	}
	now := time.Now().UTC()
	if _, err = tx.ExecContext(ctx, `UPDATE platform_subscriptions SET auto_renew_consent_at=$2,
 auto_renew_consent_expires_at=$3,auto_renew_revoked_at=NULL,
 consent_reference=$4,updated_at=$2 WHERE business_id=$1`, businessID, now, now.AddDate(1, 0, 0),
		fmt.Sprintf("checkbox:v1:user:%d:%d", userID, now.UnixNano())); err != nil {
		return nil, err
	}
	random := make([]byte, 12)
	if _, err = rand.Read(random); err != nil {
		return nil, err
	}
	reference := "platform_" + fmt.Sprint(invoice.ID) + "_" + hex.EncodeToString(random)
	var checkoutID int64
	if err = tx.GetContext(ctx, &checkoutID, `INSERT INTO platform_checkout_sessions(invoice_id,reference_id,status)
 VALUES($1,$2,'creating') RETURNING id`, invoice.ID, reference); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	returnURL := s.returnURL + "/dashboard/tagihan-platform"
	response, err := s.provider.CreateHostedSession(ctx, provider.HostedSessionRequest{
		ReferenceID: reference, Amount: invoice.Amount, BusinessID: businessID,
		PayerName: payer.Name, PayerEmail: payer.Email, SuccessURL: returnURL, CancelURL: returnURL,
		SavePaymentMethod: true,
	})
	if err != nil {
		if provider.SessionCreationRejected(err) {
			_, _ = s.db.ExecContext(context.WithoutCancel(ctx), `UPDATE platform_checkout_sessions SET status='failed',last_error='provider_rejected',updated_at=(NOW() AT TIME ZONE 'UTC') WHERE id=$1`, checkoutID)
		}
		return nil, fmt.Errorf("create platform Xendit checkout: %w", err)
	}
	if _, err = s.db.ExecContext(context.WithoutCancel(ctx), `UPDATE platform_checkout_sessions
 SET provider_session_id=$2,payment_link_url=$3,status='active',updated_at=(NOW() AT TIME ZONE 'UTC')
 WHERE id=$1 AND status='creating'`, checkoutID, response.PaymentSessionID, response.PaymentLinkURL); err != nil {
		return nil, err
	}
	return &Checkout{invoice.ID, invoice.Amount, response.PaymentLinkURL, "active"}, nil
}

func (s *Service) Cancel(ctx context.Context, businessID string) error {
	result, err := s.db.ExecContext(ctx, `UPDATE platform_subscriptions SET status='cancelled',auto_renew_revoked_at=(NOW() AT TIME ZONE 'UTC'),
 updated_at=(NOW() AT TIME ZONE 'UTC') WHERE business_id=$1 AND status IN ('active','past_due')`, businessID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return utils.ErrConflict
	}
	return nil
}

type sessionEvent struct {
	Event string `json:"event"`
	Data  struct {
		PaymentSessionID string `json:"payment_session_id"`
		ReferenceID      string `json:"reference_id"`
		Status           string `json:"status"`
		Currency         string `json:"currency"`
		Amount           int64  `json:"amount"`
		PaymentTokenID   string `json:"payment_token_id"`
		PaymentID        string `json:"payment_id"`
	} `json:"data"`
}

func (s *Service) Webhook(ctx context.Context, callbackToken string, raw []byte) error {
	if !s.provider.VerifyCallbackToken(callbackToken) {
		return utils.ErrUnauthorized
	}
	var event sessionEvent
	if len(raw) > 1<<20 || json.Unmarshal(raw, &event) != nil || event.Data.ReferenceID == "" || event.Data.PaymentSessionID == "" {
		return utils.ErrBadRequest
	}
	if event.Event != "payment_session.completed" && event.Event != "payment_session.expired" {
		return utils.ErrBadRequest
	}
	return s.applySessionEvent(ctx, event)
}

func (s *Service) applySessionEvent(ctx context.Context, event sessionEvent) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var row struct {
		ID            int64     `db:"id"`
		InvoiceID     int64     `db:"invoice_id"`
		BusinessID    string    `db:"business_id"`
		SessionID     *string   `db:"provider_session_id"`
		Amount        int64     `db:"amount"`
		SessionStatus string    `db:"session_status"`
		InvoiceStatus string    `db:"invoice_status"`
		PeriodStart   time.Time `db:"period_start"`
		PeriodEnd     time.Time `db:"period_end"`
	}
	err = tx.GetContext(ctx, &row, `SELECT cs.id,cs.invoice_id,i.business_id,cs.provider_session_id,i.amount,
 cs.status AS session_status,i.status AS invoice_status,i.period_start,i.period_end
 FROM platform_checkout_sessions cs JOIN platform_invoices i ON i.id=cs.invoice_id
 WHERE cs.reference_id=$1 FOR UPDATE OF cs,i`, event.Data.ReferenceID)
	if err == sql.ErrNoRows {
		return utils.ErrNotFound
	}
	if err != nil {
		return err
	}
	if row.SessionID != nil && *row.SessionID != event.Data.PaymentSessionID {
		return utils.ErrBadRequest
	}
	if event.Data.Currency != "IDR" || event.Data.Amount != row.Amount {
		return utils.ErrBadRequest
	}
	if row.SessionStatus == "completed" {
		return nil
	}
	if event.Event == "payment_session.completed" {
		if event.Data.Status != "COMPLETED" || event.Data.PaymentID == "" {
			return utils.ErrBadRequest
		}
		if row.InvoiceStatus != "pending" {
			return utils.ErrConflict
		}
		if _, err = tx.ExecContext(ctx, `UPDATE platform_invoices SET status='paid',paid_at=(NOW() AT TIME ZONE 'UTC') WHERE id=$1`, row.InvoiceID); err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, `UPDATE platform_subscriptions SET status=CASE WHEN status='cancelled' THEN 'cancelled' ELSE 'active' END,billing_anchor=COALESCE(billing_anchor,$2),
 next_billing_date=$3,paid_through=$3,payment_token_id=CASE WHEN auto_renew_revoked_at IS NULL AND $4<>'' THEN $4 ELSE payment_token_id END,
 updated_at=(NOW() AT TIME ZONE 'UTC') WHERE business_id=$1`, row.BusinessID, row.PeriodStart, row.PeriodEnd, event.Data.PaymentTokenID); err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `UPDATE platform_checkout_sessions SET status='completed',provider_session_id=COALESCE(provider_session_id,$2),updated_at=(NOW() AT TIME ZONE 'UTC') WHERE id=$1`, row.ID, event.Data.PaymentSessionID)
	} else {
		if event.Data.Status != "EXPIRED" {
			return utils.ErrBadRequest
		}
		_, err = tx.ExecContext(ctx, `UPDATE platform_checkout_sessions SET status='expired',provider_session_id=COALESCE(provider_session_id,$2),updated_at=(NOW() AT TIME ZONE 'UTC') WHERE id=$1 AND status IN ('creating','active')`, row.ID, event.Data.PaymentSessionID)
	}
	if err != nil {
		return err
	}
	return tx.Commit()
}

func isDefinitive(err error) bool {
	var validation *provider.ValidationError
	var api *provider.XenditAPIError
	return errors.As(err, &validation) || (errors.As(err, &api) && (api.StatusCode == 400 || api.StatusCode == 401 || api.StatusCode == 403 || api.StatusCode == 404 || api.StatusCode == 422))
}
