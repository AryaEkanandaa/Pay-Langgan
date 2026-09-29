package payment

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/mail"
	"net/url"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"pay-langgan/internal/database"
	"pay-langgan/internal/models"
	paymentrepo "pay-langgan/internal/repositories/payment"
	"pay-langgan/internal/utils"
	provider "pay-langgan/pkg/payment"
)

type SessionProvider interface {
	CreateCardSession(context.Context, provider.XenditSessionRequest) (*provider.XenditSessionResponse, error)
	GetCardSession(context.Context, string, string) (*provider.XenditSessionResponse, error)
	VerifyCallbackToken(string) bool
}

type CheckoutService struct {
	db             *database.DB
	repo           *paymentrepo.CheckoutRepository
	provider       SessionProvider
	publicBaseURL  string
	returnBaseURL  string
	ttl            time.Duration
	autopayEnabled bool
}

func NewCheckoutService(db *database.DB, repo *paymentrepo.CheckoutRepository, sessionProvider SessionProvider, publicBaseURL, returnBaseURL string, ttl time.Duration) (*CheckoutService, error) {
	publicBaseURL = strings.TrimRight(strings.TrimSpace(publicBaseURL), "/")
	returnBaseURL = strings.TrimRight(strings.TrimSpace(returnBaseURL), "/")
	if db == nil || repo == nil || sessionProvider == nil || ttl <= 0 {
		return nil, fmt.Errorf("checkout service dependencies and positive TTL are required")
	}
	if !validBaseURL(publicBaseURL, false) {
		return nil, fmt.Errorf("invalid public app URL")
	}
	if !validBaseURL(returnBaseURL, true) {
		return nil, fmt.Errorf("Xendit return base URL must use HTTPS")
	}
	return &CheckoutService{db: db, repo: repo, provider: sessionProvider,
		publicBaseURL: publicBaseURL, returnBaseURL: returnBaseURL, ttl: ttl}, nil
}

func validBaseURL(value string, requireHTTPS bool) bool {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.ForceQuery || parsed.Fragment != "" {
		return false
	}
	if requireHTTPS {
		return parsed.Scheme == "https"
	}
	return parsed.Scheme == "http" || parsed.Scheme == "https"
}

func (s *CheckoutService) CreateLink(ctx context.Context, invoiceID int, businessID string) (*models.CheckoutLinkResponse, error) {
	if invoiceID < 1 || businessID == "" {
		return nil, fmt.Errorf("%w: invoice and business are required", utils.ErrBadRequest)
	}
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin checkout link transaction: %w", err)
	}
	defer tx.Rollback()
	record, err := s.repo.FindInvoiceForCheckout(tx, invoiceID, businessID)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, utils.ErrNotFound
	}
	amount := int64(record.InvoiceAmount)
	if record.InvoiceStatus != "pending" {
		return nil, fmt.Errorf("%w: only pending invoices can be paid", utils.ErrConflict)
	}
	if float64(amount) != record.InvoiceAmount || amount < 5000 || amount > 200000000 {
		return nil, fmt.Errorf("%w: Xendit card checkout requires a whole-rupiah amount between IDR 5,000 and IDR 200,000,000", utils.ErrBadRequest)
	}
	var activeSessions int
	if err := tx.GetContext(ctx, &activeSessions, `SELECT count(*) FROM checkout_sessions
 WHERE invoice_id = $1 AND status IN ('creating', 'active')`, invoiceID); err != nil {
		return nil, fmt.Errorf("check active checkout: %w", err)
	}
	if activeSessions > 0 {
		return nil, fmt.Errorf("%w: invoice already has an active checkout", utils.ErrConflict)
	}
	if err := tx.GetContext(ctx, &activeSessions, `SELECT count(*) FROM autopay_charges WHERE invoice_id=$1 AND status IN ('creating','pending','review_required')`, invoiceID); err != nil {
		return nil, err
	}
	if activeSessions > 0 {
		return nil, fmt.Errorf("%w: automatic payment is in progress or awaiting reconciliation", utils.ErrConflict)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE checkout_sessions SET status = 'expired',
 updated_at = (NOW() AT TIME ZONE 'UTC')
 WHERE invoice_id = $1 AND status IN ('created', 'failed')`, invoiceID); err != nil {
		return nil, fmt.Errorf("expire previous checkout links: %w", err)
	}
	token, err := secureToken(32)
	if err != nil {
		return nil, fmt.Errorf("generate checkout token: %w", err)
	}
	suffix, err := secureToken(12)
	if err != nil {
		return nil, fmt.Errorf("generate checkout reference: %w", err)
	}
	checkout := &models.CheckoutSession{
		InvoiceID:       invoiceID,
		PublicTokenHash: hashToken(token),
		ReferenceID:     fmt.Sprintf("invoice_%d_%s", invoiceID, suffix),
		ExpiresAt:       time.Now().UTC().Add(s.ttl),
	}
	if err := s.repo.Create(tx, checkout); err != nil {
		return nil, fmt.Errorf("create checkout link: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit checkout link: %w", err)
	}
	return &models.CheckoutLinkResponse{
		Token: token, CheckoutURL: s.publicBaseURL + "/checkout/" + token, ExpiresAt: checkout.ExpiresAt,
	}, nil
}

func (s *CheckoutService) Get(ctx context.Context, token string) (*models.PublicCheckoutResponse, error) {
	record, err := s.findByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	status := record.Status
	if time.Now().UTC().After(record.ExpiresAt) && (status == "created" || status == "failed") {
		status = "expired"
	}
	return &models.PublicCheckoutResponse{
		InvoiceNumber: record.InvoiceNumber, Amount: record.InvoiceAmount,
		InvoiceStatus: record.InvoiceStatus, CheckoutStatus: status,
		CustomerName: record.CustomerName, CustomerEmail: record.DefaultEmail,
		CustomerMobile: record.DefaultMobile, BusinessName: record.BusinessName,
		PlanName: record.PlanName, ExpiresAt: record.ExpiresAt,
	}, nil
}

func (s *CheckoutService) Start(ctx context.Context, token string, request models.StartCheckoutRequest) (*models.StartCheckoutResponse, error) {
	record, err := s.findByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if record.InvoiceStatus != "pending" || !record.ExpiresAt.After(time.Now()) {
		return nil, fmt.Errorf("%w: checkout is no longer available", utils.ErrConflict)
	}
	if record.Status == "active" && record.ProviderSessionID != nil {
		if record.PaymentLinkURL != nil && *record.PaymentLinkURL != "" {
			return nil, fmt.Errorf("%w: existing hosted session must finish or expire before using components", utils.ErrConflict)
		}
		session, err := s.provider.GetCardSession(ctx, *record.ProviderSessionID, record.ReferenceID)
		if err != nil {
			return nil, fmt.Errorf("retrieve Xendit checkout: %w", err)
		}
		return componentsCheckoutResponse(session, record.ExpiresAt)
	}
	if record.InvoiceStatus != "pending" || record.Status != "created" || time.Now().UTC().After(record.ExpiresAt) {
		return nil, fmt.Errorf("%w: checkout is no longer available", utils.ErrConflict)
	}
	name := strings.TrimSpace(request.Name)
	if name == "" {
		name = record.CustomerName
	}
	email := strings.TrimSpace(request.Email)
	mobile := strings.TrimSpace(request.Mobile)
	if err := validateCheckoutContact(name, email, mobile); err != nil {
		return nil, err
	}
	returnURL := s.returnBaseURL + "/checkout/" + token
	if request.ParentOrigin != "" {
		parent, err := url.Parse(request.ParentOrigin)
		if err != nil || parent.Host == "" || parent.String() != parent.Scheme+"://"+parent.Host ||
			(parent.Scheme != "https" && !(parent.Scheme == "http" && (parent.Hostname() == "localhost" || parent.Hostname() == "127.0.0.1"))) {
			return nil, fmt.Errorf("%w: invalid parent origin", utils.ErrBadRequest)
		}
		returnURL += "?parent_origin=" + url.QueryEscape(request.ParentOrigin)
	}
	if claimed, err := s.repo.Claim(ctx, record.ID, email, mobile); err != nil {
		return nil, err
	} else if !claimed {
		return nil, fmt.Errorf("%w: checkout is already being processed", utils.ErrConflict)
	}
	amount := int64(record.InvoiceAmount)
	appURL, _ := url.Parse(s.publicBaseURL)
	response, err := s.provider.CreateCardSession(ctx, provider.XenditSessionRequest{
		ReferenceID: record.ReferenceID, Amount: amount,
		CustomerReferenceID: fmt.Sprintf("customer_%d", record.CustomerID), CustomerName: name,
		CustomerEmail: email, CustomerMobile: mobile,
		Description:       fmt.Sprintf("%s - %s", record.InvoiceNumber, record.PlanName),
		ComponentsOrigins: []string{appURL.Scheme + "://" + appURL.Host},
		ReturnURL:         returnURL,
	})
	if err != nil {
		// A disconnected browser must not prevent recording an uncertain outcome.
		recoveryCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		if provider.SessionCreationRejected(err) {
			if saveErr := s.repo.MarkFailed(recoveryCtx, record.ID); saveErr != nil {
				return nil, saveErr
			}
			var validation *provider.ValidationError
			if errors.As(err, &validation) {
				return nil, fmt.Errorf("%w: %v", utils.ErrBadRequest, validation)
			}
		} else {
			_, saveErr := s.db.ExecContext(recoveryCtx, `UPDATE checkout_sessions SET last_error='session_creation_uncertain', review_required=TRUE,
 next_reconcile_at=(NOW() AT TIME ZONE 'UTC') + INTERVAL '1 minute' WHERE id=$1 AND status='creating'`, record.ID)
			if saveErr != nil {
				return nil, saveErr
			}
		}
		return nil, fmt.Errorf("create Xendit checkout: %w", err)
	}
	recoveryCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	if err := s.repo.RememberSession(recoveryCtx, record.ID, response.PaymentSessionID); err != nil {
		return nil, err
	}
	checkout, err := componentsCheckoutResponse(response, record.ExpiresAt)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Activate(recoveryCtx, record.ID, record.InvoiceID, record.InvoiceAmount, response.PaymentSessionID, checkout.ExpiresAt); err != nil {
		return nil, err
	}
	return checkout, nil
}

func componentsCheckoutResponse(session *provider.XenditSessionResponse, expiresAt time.Time) (*models.StartCheckoutResponse, error) {
	if session == nil || session.Mode != "COMPONENTS" || session.Status != "ACTIVE" ||
		session.PaymentSessionID == "" || strings.TrimSpace(session.ComponentsSDKKey) == "" || !session.ExpiresAt.After(time.Now()) || !expiresAt.After(time.Now()) {
		return nil, fmt.Errorf("%w: components checkout is no longer available", utils.ErrConflict)
	}
	if session.ExpiresAt.Before(expiresAt) {
		expiresAt = session.ExpiresAt
	}
	return &models.StartCheckoutResponse{PaymentSessionID: session.PaymentSessionID, Mode: "COMPONENTS",
		ComponentsSDKKey: session.ComponentsSDKKey, Status: "active", ExpiresAt: expiresAt}, nil
}

func validateCheckoutContact(name, email, mobile string) error {
	if name == "" || len(name) > 100 {
		return fmt.Errorf("%w: customer name is required and must be at most 100 characters", utils.ErrBadRequest)
	}
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email || len(email) > 254 {
		return fmt.Errorf("%w: valid customer email is required", utils.ErrBadRequest)
	}
	if len(mobile) < 9 || len(mobile) > 16 || mobile[0] != '+' {
		return fmt.Errorf("%w: mobile number must use E.164 format", utils.ErrBadRequest)
	}
	for _, char := range mobile[1:] {
		if char < '0' || char > '9' {
			return fmt.Errorf("%w: mobile number must use E.164 format", utils.ErrBadRequest)
		}
	}
	return nil
}

func (s *CheckoutService) findByToken(ctx context.Context, token string) (*models.CheckoutRecord, error) {
	if len(token) < 32 || len(token) > 128 {
		return nil, utils.ErrNotFound
	}
	record, err := s.repo.FindByTokenHash(ctx, hashToken(token))
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, utils.ErrNotFound
	}
	return record, nil
}

type xenditSessionWebhook struct {
	Event string `json:"event"`
	Data  struct {
		PaymentSessionID string `json:"payment_session_id"`
		ReferenceID      string `json:"reference_id"`
		Status           string `json:"status"`
		Currency         string `json:"currency"`
		Amount           int64  `json:"amount"`
		PaymentTokenID   string `json:"payment_token_id"`
		PaymentRequestID string `json:"payment_request_id"`
		PaymentID        string `json:"payment_id"`
	} `json:"data"`
}

func (s *CheckoutService) ProcessWebhook(ctx context.Context, callbackToken string, payload []byte) (bool, error) {
	if !s.provider.VerifyCallbackToken(callbackToken) {
		return false, utils.ErrUnauthorized
	}
	var event xenditSessionWebhook
	if err := json.Unmarshal(payload, &event); err != nil {
		return false, fmt.Errorf("%w: invalid webhook payload", utils.ErrBadRequest)
	}
	if event.Event != "payment_session.completed" && event.Event != "payment_session.expired" {
		return false, fmt.Errorf("%w: unsupported Xendit event", utils.ErrBadRequest)
	}
	if event.Data.PaymentSessionID == "" || event.Data.ReferenceID == "" {
		return false, fmt.Errorf("%w: incomplete Xendit session event", utils.ErrBadRequest)
	}

	return s.acceptSessionEvent(ctx, event)
}

// Called only for authenticated inbox events or verified provider GET responses.
// The caller owns the transaction; failed effects are rolled back to a savepoint.
func (s *CheckoutService) applySessionEvent(ctx context.Context, tx *sqlx.Tx, event xenditSessionWebhook) error {
	var checkout struct {
		ID          int64   `db:"id"`
		InvoiceID   int     `db:"invoice_id"`
		PaymentID   *int    `db:"payment_id"`
		CustomerID  int     `db:"customer_id"`
		ReferenceID string  `db:"reference_id"`
		Amount      float64 `db:"amount"`
		Invoice     string  `db:"invoice_status"`
		Status      string  `db:"checkout_status"`
		SessionID   *string `db:"provider_session_id"`
	}
	var invoiceID int
	if err := tx.GetContext(ctx, &invoiceID, `SELECT invoice_id FROM checkout_sessions WHERE reference_id=$1`, event.Data.ReferenceID); err != nil {
		return fmt.Errorf("checkout not found: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `SELECT id FROM invoices WHERE id=$1 FOR UPDATE`, invoiceID); err != nil {
		return err
	}
	err := tx.GetContext(ctx, &checkout, `SELECT cs.id, cs.invoice_id, cs.payment_id, cs.provider_session_id,
 c.id AS customer_id, cs.reference_id, i.amount::float8 AS amount, i.status AS invoice_status,
 cs.status AS checkout_status
 FROM checkout_sessions cs JOIN invoices i ON i.id = cs.invoice_id
 JOIN subscriptions s ON s.id = i.subscription_id JOIN customers c ON c.id = s.customer_id
 WHERE cs.reference_id = $1 FOR UPDATE OF cs`, event.Data.ReferenceID)
	if err == sql.ErrNoRows {
		return fmt.Errorf("checkout session was not found")
	}
	if err != nil {
		return fmt.Errorf("lock checkout for webhook: %w", err)
	}
	if checkout.ReferenceID != event.Data.ReferenceID || event.Data.Currency != "IDR" ||
		math.Abs(checkout.Amount-float64(event.Data.Amount)) > 0.0001 {
		return fmt.Errorf("%w: webhook does not match the checkout", utils.ErrBadRequest)
	}
	if checkout.SessionID != nil && *checkout.SessionID != event.Data.PaymentSessionID {
		return fmt.Errorf("%w: provider session mismatch", utils.ErrBadRequest)
	}
	if event.Event == "payment_session.completed" && (event.Data.Status != "COMPLETED" || event.Data.PaymentID == "") {
		return fmt.Errorf("%w: incomplete completed session", utils.ErrBadRequest)
	}
	if event.Event == "payment_session.expired" && event.Data.Status != "EXPIRED" {
		return fmt.Errorf("%w: incomplete expired session", utils.ErrBadRequest)
	}
	// The webhook can arrive before POST /sessions returns, or recover a lost
	// creation response. Attach by our unique reference only while creating.
	if checkout.Status == "creating" {
		if checkout.PaymentID == nil {
			var paymentID int
			if err := tx.QueryRowContext(ctx, `INSERT INTO payments(invoice_id,provider,provider_transaction_id,amount,status,created_at)
 VALUES($1,'xendit',$2,$3,'pending',(NOW() AT TIME ZONE 'UTC')) RETURNING id`, checkout.InvoiceID, event.Data.PaymentSessionID, checkout.Amount).Scan(&paymentID); err != nil {
				return err
			}
			checkout.PaymentID = &paymentID
		}
		if _, err := tx.ExecContext(ctx, `UPDATE checkout_sessions SET provider_session_id=$2,payment_id=$3,status='active',
 review_required=FALSE,last_error=NULL,updated_at=(NOW() AT TIME ZONE 'UTC') WHERE id=$1`, checkout.ID, event.Data.PaymentSessionID, *checkout.PaymentID); err != nil {
			return err
		}
		checkout.Status = "active"
	}
	if checkout.PaymentID == nil {
		return fmt.Errorf("checkout has no pending payment")
	}
	// Completion is terminal. A late expiration must never downgrade it.
	if checkout.Status == "completed" {
		if event.Event == "payment_session.completed" {
			var paymentID string
			if err := tx.GetContext(ctx, &paymentID, `SELECT provider_transaction_id FROM payments WHERE id=$1`, *checkout.PaymentID); err != nil {
				return err
			}
			if paymentID != event.Data.PaymentID {
				return fmt.Errorf("%w: payment identity mismatch", utils.ErrBadRequest)
			}
		}
		return nil
	}

	if event.Event == "payment_session.completed" {
		if checkout.Status != "active" && checkout.Status != "expired" {
			return fmt.Errorf("%w: checkout is not active", utils.ErrConflict)
		}
		result, err := tx.ExecContext(ctx, `UPDATE invoices SET status = 'paid', paid_at = COALESCE(paid_at, (NOW() AT TIME ZONE 'UTC'))
 WHERE id = $1 AND status IN ('pending', 'paid')`, checkout.InvoiceID)
		if err != nil {
			return fmt.Errorf("mark invoice paid from webhook: %w", err)
		}
		rows, _ := result.RowsAffected()
		if rows != 1 {
			return fmt.Errorf("%w: invoice cannot accept payment", utils.ErrConflict)
		}
		_, err = tx.ExecContext(ctx, `UPDATE payments SET provider_transaction_id = $2,
 status = 'success', paid_at = (NOW() AT TIME ZONE 'UTC') WHERE id = $1`, *checkout.PaymentID, event.Data.PaymentID)
		if err != nil {
			return fmt.Errorf("complete payment: %w", err)
		}
		if event.Data.PaymentTokenID != "" {
			_, err = tx.ExecContext(ctx, `INSERT INTO payment_methods (customer_id, provider, token, is_default, created_at)
 VALUES ($1, 'xendit', $2, NOT EXISTS (SELECT 1 FROM payment_methods WHERE customer_id = $1), (NOW() AT TIME ZONE 'UTC'))
 ON CONFLICT (provider, token) DO NOTHING`, checkout.CustomerID, event.Data.PaymentTokenID)
			if err != nil {
				return fmt.Errorf("save Xendit payment token: %w", err)
			}
		}
		_, err = tx.ExecContext(ctx, `UPDATE checkout_sessions SET status = 'completed',review_required=FALSE,last_error=NULL, updated_at = (NOW() AT TIME ZONE 'UTC') WHERE id = $1 AND status IN ('active','expired')`, checkout.ID)
		if err != nil {
			return fmt.Errorf("complete checkout: %w", err)
		}
	} else {
		_, err = tx.ExecContext(ctx, `UPDATE checkout_sessions SET status = 'expired',review_required=FALSE,last_error=NULL, updated_at = (NOW() AT TIME ZONE 'UTC') WHERE id = $1 AND status = 'active'`, checkout.ID)
		if err == nil {
			_, err = tx.ExecContext(ctx, `UPDATE payments SET status = 'expired' WHERE id = $1 AND status = 'pending'`, *checkout.PaymentID)
		}
		if err != nil {
			return fmt.Errorf("expire checkout: %w", err)
		}
	}
	return nil
}

func secureToken(size int) (string, error) {
	buffer := make([]byte, size)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
