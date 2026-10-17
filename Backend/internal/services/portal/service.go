package portal

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"net/mail"
	"net/url"
	"strings"
	"time"

	"pay-langgan/internal/config"
	"pay-langgan/internal/database"
	"pay-langgan/internal/models"
	"pay-langgan/internal/utils"
)

type Service struct {
	db  *database.DB
	cfg *config.Config
}

func NewService(db *database.DB, cfg *config.Config) *Service { return &Service{db: db, cfg: cfg} }

type Business struct {
	ID     string `db:"business_id" json:"-"`
	Name   string `db:"name" json:"name"`
	SiteID int64  `db:"site_id" json:"-"`
}

func (s *Service) Business(ctx context.Context, site string) (*Business, error) {
	var b Business
	err := s.db.GetContext(ctx, &b, `SELECT b.id business_id,b.name,es.id site_id FROM merchant_embed_sites es JOIN businesses b ON b.id=es.business_id WHERE es.public_id=$1 AND es.revoked_at IS NULL AND b.status='active' AND NOT b.deleted`, site)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, utils.ErrNotFound
	}
	return &b, err
}
func magicToken(secret string, nonce []byte) string {
	m := hmac.New(sha256.New, []byte(secret))
	m.Write([]byte("paylanggan-customer-portal-v1:"))
	m.Write(nonce)
	return base64.RawURLEncoding.EncodeToString(m.Sum(nil))
}
func tokenHash(token string) []byte { h := sha256.Sum256([]byte(token)); return h[:] }
func validToken(token string) bool {
	b, e := base64.RawURLEncoding.DecodeString(token)
	return e == nil && len(b) == 32 && len(token) == 43
}
func (s *Service) baseURL() (string, error) {
	u, e := url.Parse(s.cfg.AppPublicURL)
	if e != nil || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return "", fmt.Errorf("invalid APP_PUBLIC_URL")
	}
	local := u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1"
	if u.Scheme != "https" && !(s.cfg.EmailDeliveryMode == "preview" && local && u.Scheme == "http") {
		return "", fmt.Errorf("portal email requires HTTPS APP_PUBLIC_URL")
	}
	return strings.TrimRight(u.String(), "/"), nil
}

// The response never discloses whether this email belongs to a customer.
func (s *Service) RequestLink(ctx context.Context, site, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	a, e := mail.ParseAddress(email)
	if e != nil || a.Address != email || len(email) > 100 {
		return utils.ErrBadRequest
	}
	if s.cfg.EmailDeliveryMode != "smtp" && s.cfg.EmailDeliveryMode != "preview" {
		return fmt.Errorf("portal email unavailable")
	}
	if _, e = s.baseURL(); e != nil {
		return e
	}
	b, e := s.Business(ctx, site)
	if e != nil {
		return e
	}
	tx, e := s.db.BeginTxx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	// Serialize per email even when checkout produced duplicate customer rows.
	if _, e = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, b.ID+":"+email); e != nil {
		return e
	}
	var exists bool
	if e = tx.GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM customers WHERE business_id=$1 AND lower(trim(email))=$2)`, b.ID, email); e != nil {
		return e
	}
	if !exists {
		return nil
	}
	var recent int
	var cooldown bool
	if e = tx.QueryRowContext(ctx, `SELECT COUNT(*),COALESCE(bool_or(created_at>NOW()-INTERVAL '1 minute'),FALSE) FROM customer_portal_requests WHERE site_id=$1 AND recipient=$2 AND created_at>NOW()-INTERVAL '1 hour'`, b.SiteID, email).Scan(&recent, &cooldown); e != nil {
		return e
	}
	if cooldown || recent >= 5 {
		return nil
	}
	nonce := make([]byte, 32)
	if _, e = rand.Read(nonce); e != nil {
		return e
	}
	_, e = tx.ExecContext(ctx, `INSERT INTO customer_portal_requests(site_id,recipient,token_hash,nonce) VALUES($1,$2,$3,$4)`, b.SiteID, email, tokenHash(magicToken(s.cfg.JWTSecret, nonce)), nonce)
	if e != nil {
		return e
	}
	return tx.Commit()
}

type Login struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (s *Service) Exchange(ctx context.Context, site, token string) (*Login, error) {
	if !validToken(token) {
		return nil, utils.ErrUnauthorized
	}
	b, e := s.Business(ctx, site)
	if e != nil {
		return nil, e
	}
	tx, e := s.db.BeginTxx(ctx, nil)
	if e != nil {
		return nil, e
	}
	defer tx.Rollback()
	var email string
	e = tx.QueryRowContext(ctx, `UPDATE customer_portal_requests r SET used_at=NOW(),nonce=NULL,status=CASE WHEN status IN ('queued','sending') THEN 'cancelled' ELSE status END WHERE r.site_id=$1 AND r.token_hash=$2 AND r.used_at IS NULL AND r.expires_at>NOW() AND EXISTS(SELECT 1 FROM customers c WHERE c.business_id=$3 AND lower(trim(c.email))=r.recipient) RETURNING recipient`, b.SiteID, tokenHash(token), b.ID).Scan(&email)
	if errors.Is(e, sql.ErrNoRows) {
		return nil, utils.ErrUnauthorized
	}
	if e != nil {
		return nil, e
	}
	raw := make([]byte, 32)
	if _, e = rand.Read(raw); e != nil {
		return nil, e
	}
	result := &Login{Token: base64.RawURLEncoding.EncodeToString(raw)}
	e = tx.QueryRowContext(ctx, `INSERT INTO customer_portal_sessions(site_id,recipient,token_hash) VALUES($1,$2,$3) RETURNING expires_at`, b.SiteID, email, tokenHash(result.Token)).Scan(&result.ExpiresAt)
	if e != nil {
		return nil, e
	}
	if e = tx.Commit(); e != nil {
		return nil, e
	}
	return result, nil
}

type Identity struct {
	BusinessID   string `db:"business_id"`
	BusinessName string `db:"business_name"`
	Email        string `db:"recipient"`
}

func (s *Service) Authenticate(ctx context.Context, site, token string) (*Identity, error) {
	if !validToken(token) {
		return nil, utils.ErrUnauthorized
	}
	var v Identity
	e := s.db.GetContext(ctx, &v, `SELECT b.id business_id,b.name business_name,p.recipient FROM customer_portal_sessions p JOIN merchant_embed_sites es ON es.id=p.site_id JOIN businesses b ON b.id=es.business_id WHERE p.token_hash=$1 AND p.expires_at>NOW() AND es.public_id=$2 AND es.revoked_at IS NULL AND b.status='active' AND NOT b.deleted AND EXISTS(SELECT 1 FROM customers c WHERE c.business_id=b.id AND lower(trim(c.email))=p.recipient)`, tokenHash(token), site)
	if errors.Is(e, sql.ErrNoRows) {
		return nil, utils.ErrUnauthorized
	}
	return &v, e
}
func (s *Service) Logout(ctx context.Context, token string) error {
	_, e := s.db.ExecContext(ctx, `DELETE FROM customer_portal_sessions WHERE token_hash=$1`, tokenHash(token))
	return e
}

type Subscription struct {
	ID        int        `db:"id" json:"id"`
	Plan      string     `db:"plan" json:"plan"`
	Status    string     `db:"status" json:"status"`
	Price     float64    `db:"price" json:"price"`
	Cycle     string     `db:"cycle" json:"cycle"`
	Start     time.Time  `db:"start_date" json:"start_date"`
	Next      *time.Time `db:"next_billing_date" json:"next_billing_date"`
	TrialEnds *time.Time `db:"trial_ends_at" json:"trial_ends_at"`
}
type Overview struct {
	BusinessName      string                   `json:"business_name"`
	Email             string                   `json:"email"`
	Subscriptions     []Subscription           `json:"subscriptions"`
	Invoices          []models.InvoiceListItem `json:"invoices"`
	SubscriptionTotal int                      `json:"subscription_total"`
	InvoiceTotal      int                      `json:"invoice_total"`
}

const customerScope = `c.business_id=$1 AND lower(trim(c.email))=$2`

func (s *Service) Overview(ctx context.Context, v *Identity, subPage, invoicePage int) (*Overview, error) {
	if subPage < 1 || invoicePage < 1 || subPage > 100000 || invoicePage > 100000 {
		return nil, utils.ErrBadRequest
	}
	o := &Overview{BusinessName: v.BusinessName, Email: v.Email, Subscriptions: []Subscription{}, Invoices: []models.InvoiceListItem{}}
	subFrom := ` FROM subscriptions s JOIN customers c ON c.id=s.customer_id JOIN plans p ON p.id=s.plan_id WHERE ` + customerScope
	if e := s.db.GetContext(ctx, &o.SubscriptionTotal, `SELECT COUNT(*)`+subFrom, v.BusinessID, v.Email); e != nil {
		return nil, e
	}
	if e := s.db.SelectContext(ctx, &o.Subscriptions, `SELECT s.id,p.name plan,s.status,p.price,COALESCE(s.billing_cycle,p.billing_cycle) cycle,s.start_date,s.next_billing_date,s.trial_ends_at`+subFrom+` ORDER BY s.id DESC LIMIT 20 OFFSET $3`, v.BusinessID, v.Email, (subPage-1)*20); e != nil {
		return nil, e
	}
	invFrom := ` FROM invoices i JOIN subscriptions s ON s.id=i.subscription_id JOIN customers c ON c.id=s.customer_id JOIN plans p ON p.id=s.plan_id WHERE ` + customerScope
	if e := s.db.GetContext(ctx, &o.InvoiceTotal, `SELECT COUNT(*)`+invFrom, v.BusinessID, v.Email); e != nil {
		return nil, e
	}
	e := s.db.SelectContext(ctx, &o.Invoices, `SELECT i.id,i.subscription_id,i.invoice_number,i.amount,i.status,i.due_date,i.paid_at,i.created_at,i.period_start,i.period_end,c.id customer_id,c.name customer_name,c.email customer_email,p.name plan_name`+invFrom+` ORDER BY i.id DESC LIMIT 20 OFFSET $3`, v.BusinessID, v.Email, (invoicePage-1)*20)
	return o, e
}
func (s *Service) OwnInvoice(ctx context.Context, v *Identity, id int) error {
	var exists bool
	e := s.db.GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM invoices i JOIN subscriptions s ON s.id=i.subscription_id JOIN customers c ON c.id=s.customer_id WHERE i.id=$3 AND `+customerScope+`)`, v.BusinessID, v.Email, id)
	if e != nil {
		return e
	}
	if !exists {
		return utils.ErrNotFound
	}
	return nil
}
