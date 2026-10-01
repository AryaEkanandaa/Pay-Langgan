package payment

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"pay-langgan/internal/database"
	billingrepo "pay-langgan/internal/repositories/billing"
	billingsvc "pay-langgan/internal/services/billing"
	"pay-langgan/internal/utils"
)

type MerchantIntegrationService struct {
	db       *database.DB
	checkout *CheckoutService
	invoices *billingrepo.InvoiceRepository
}

func NewMerchantIntegrationService(db *database.DB, checkout *CheckoutService, invoices *billingrepo.InvoiceRepository) *MerchantIntegrationService {
	return &MerchantIntegrationService{db: db, checkout: checkout, invoices: invoices}
}

type MerchantOrderRequest struct {
	ExternalOrderID string `json:"external_order_id"`
	PlanID          int    `json:"plan_id"`
	Customer        struct {
		Name   string `json:"name"`
		Email  string `json:"email"`
		Mobile string `json:"mobile"`
	} `json:"customer"`
}

type MerchantOrderResponse struct {
	ExternalOrderID string `db:"external_order_id" json:"external_order_id"`
	InvoiceID       int    `db:"invoice_id" json:"invoice_id"`
	InvoiceStatus   string `db:"invoice_status" json:"invoice_status"`
	CheckoutURL     string `json:"checkout_url,omitempty"`
}

type MerchantKeyInfo struct {
	ID        int64      `db:"id" json:"id"`
	Name      string     `db:"name" json:"name"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	RevokedAt *time.Time `db:"revoked_at" json:"revoked_at,omitempty"`
}

type MerchantKeyCreated struct {
	MerchantKeyInfo
	SecretKey string `json:"secret_key"`
}

func merchantKeyHash(key string) string {
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:])
}

func (s *MerchantIntegrationService) CreateKey(ctx context.Context, businessID, name string) (*MerchantKeyCreated, error) {
	name = strings.TrimSpace(name)
	if businessID == "" || name == "" || len(name) > 100 {
		return nil, utils.ErrBadRequest
	}
	random, err := secureToken(32)
	if err != nil {
		return nil, err
	}
	key := "pl_sk_" + random
	created := &MerchantKeyCreated{MerchantKeyInfo: MerchantKeyInfo{Name: name}, SecretKey: key}
	err = s.db.QueryRowContext(ctx, `INSERT INTO merchant_api_keys (business_id, name, key_hash)
 VALUES ($1,$2,$3) RETURNING id, created_at`, businessID, name, merchantKeyHash(key)).Scan(&created.ID, &created.CreatedAt)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s *MerchantIntegrationService) ListKeys(ctx context.Context, businessID string) ([]MerchantKeyInfo, error) {
	keys := []MerchantKeyInfo{}
	err := s.db.SelectContext(ctx, &keys, `SELECT id, name, created_at, revoked_at FROM merchant_api_keys
 WHERE business_id=$1 ORDER BY id DESC`, businessID)
	return keys, err
}

func (s *MerchantIntegrationService) RevokeKey(ctx context.Context, businessID string, keyID int64) (bool, error) {
	result, err := s.db.ExecContext(ctx, `UPDATE merchant_api_keys SET revoked_at=(NOW() AT TIME ZONE 'UTC')
 WHERE id=$1 AND business_id=$2 AND revoked_at IS NULL`, keyID, businessID)
	if err != nil {
		return false, err
	}
	n, err := result.RowsAffected()
	return n == 1, err
}

func (s *MerchantIntegrationService) Authenticate(ctx context.Context, key string) (string, error) {
	if !strings.HasPrefix(key, "pl_sk_") || len(key) < 40 || len(key) > 100 {
		return "", utils.ErrUnauthorized
	}
	var businessID string
	err := s.db.GetContext(ctx, &businessID, `SELECT k.business_id FROM merchant_api_keys k
 JOIN businesses b ON b.id=k.business_id
 WHERE k.key_hash=$1 AND k.revoked_at IS NULL AND b.status='active' AND b.deleted=FALSE`, merchantKeyHash(key))
	if errors.Is(err, sql.ErrNoRows) {
		return "", utils.ErrUnauthorized
	}
	return businessID, err
}

func validateMerchantOrder(req MerchantOrderRequest) (MerchantOrderRequest, error) {
	req.ExternalOrderID = strings.TrimSpace(req.ExternalOrderID)
	req.Customer.Name = strings.TrimSpace(req.Customer.Name)
	req.Customer.Email = strings.TrimSpace(req.Customer.Email)
	req.Customer.Mobile = strings.TrimSpace(req.Customer.Mobile)
	if req.PlanID < 1 || req.ExternalOrderID == "" || len(req.ExternalOrderID) > 100 ||
		strings.ContainsAny(req.ExternalOrderID, "\r\n") || len(req.Customer.Email) > 100 || len(req.Customer.Mobile) > 20 ||
		validateCheckoutContact(req.Customer.Name, req.Customer.Email, req.Customer.Mobile) != nil {
		return req, utils.ErrBadRequest
	}
	return req, nil
}

// CreateOrder atomically records an external order, customer, subscription and
// first invoice. A duplicate external_order_id can never create a second bill.
func (s *MerchantIntegrationService) CreateOrder(ctx context.Context, businessID string, request MerchantOrderRequest) (*MerchantOrderResponse, error) {
	req, err := validateMerchantOrder(request)
	if err != nil {
		return nil, err
	}
	var entitled bool
	if err := s.db.GetContext(ctx, &entitled, `SELECT EXISTS (SELECT 1 FROM platform_subscriptions
 WHERE business_id=$1 AND status='active' AND paid_through > (NOW() AT TIME ZONE 'UTC'))`, businessID); err != nil {
		return nil, err
	}
	if !entitled {
		return nil, fmt.Errorf("%w: PayLanggan business subscription is not active", utils.ErrForbidden)
	}
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var plan struct {
		Price        float64 `db:"price"`
		BillingCycle string  `db:"billing_cycle"`
		TrialDays    int     `db:"trial_days"`
	}
	err = tx.GetContext(ctx, &plan, `SELECT p.price, p.billing_cycle, p.trial_days FROM plans p
 JOIN products pr ON pr.id=p.product_id JOIN services sv ON sv.id=pr.service_id
 WHERE p.id=$1 AND sv.business_id=$2 AND pr.status='active'`, req.PlanID, businessID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, utils.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if plan.TrialDays != 0 || int64(plan.Price) < 5000 || int64(plan.Price) > 200000000 || float64(int64(plan.Price)) != plan.Price ||
		(plan.BillingCycle != "monthly" && plan.BillingCycle != "yearly") {
		return nil, fmt.Errorf("%w: this plan is not supported by card checkout", utils.ErrBadRequest)
	}
	var orderID int64
	err = tx.QueryRowContext(ctx, `INSERT INTO merchant_orders (business_id, external_order_id, plan_id, customer_email)
 VALUES ($1,$2,$3,$4) ON CONFLICT (business_id, external_order_id) DO NOTHING RETURNING id`, businessID, req.ExternalOrderID, req.PlanID, req.Customer.Email).Scan(&orderID)
	if errors.Is(err, sql.ErrNoRows) {
		var existing struct {
			PlanID        int    `db:"plan_id"`
			CustomerEmail string `db:"customer_email"`
		}
		if err := tx.GetContext(ctx, &existing, `SELECT plan_id, customer_email FROM merchant_orders
 WHERE business_id=$1 AND external_order_id=$2`, businessID, req.ExternalOrderID); err != nil {
			return nil, err
		}
		if existing.PlanID != req.PlanID || existing.CustomerEmail != req.Customer.Email {
			return nil, fmt.Errorf("%w: order reference already belongs to another purchase", utils.ErrConflict)
		}
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		return s.checkoutLinkForOrder(ctx, businessID, req.ExternalOrderID)
	}
	if err != nil {
		return nil, err
	}

	var customerID int
	// Serialize purchases for the same business and email so concurrent orders
	// reuse one customer record instead of creating more duplicates.
	if _, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext($1), hashtext(lower($2)))`, businessID, req.Customer.Email); err != nil {
		return nil, err
	}
	err = tx.QueryRowContext(ctx, `SELECT c.id FROM customers c
 WHERE c.business_id=$1 AND lower(c.email)=lower($2)
 ORDER BY EXISTS (SELECT 1 FROM payment_methods pm WHERE pm.customer_id=c.id AND pm.is_default=true) DESC, c.id
 LIMIT 1`, businessID, req.Customer.Email).Scan(&customerID)
	if errors.Is(err, sql.ErrNoRows) {
		err = tx.QueryRowContext(ctx, `INSERT INTO customers (business_id,name,email,contact,meta)
 VALUES ($1,$2,$3,$4,'{}'::jsonb) RETURNING id`, businessID, req.Customer.Name, req.Customer.Email, req.Customer.Mobile).Scan(&customerID)
	}
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	var next time.Time
	if plan.BillingCycle == "monthly" {
		next = now.AddDate(0, 1, 0)
	} else {
		next = now.AddDate(1, 0, 0)
	}
	var subscriptionID int
	err = tx.QueryRowContext(ctx, `INSERT INTO subscriptions
 (customer_id,plan_id,status,start_date,billing_anchor,billing_cycle,next_billing_date,meta)
 VALUES ($1,$2,'active',$3,$3,$4,$5,'{}'::jsonb) RETURNING id`, customerID, req.PlanID, now, plan.BillingCycle, next).Scan(&subscriptionID)
	if err != nil {
		return nil, err
	}
	invoice, err := billingsvc.IssueInvoiceContext(ctx, tx, s.invoices, subscriptionID, now, next, plan.BillingCycle, true)
	if err != nil {
		return nil, err
	}
	if invoice.Status != "pending" {
		return nil, fmt.Errorf("%w: invoice is not payable", utils.ErrBadRequest)
	}
	_, err = tx.ExecContext(ctx, `UPDATE merchant_orders SET invoice_id=$2 WHERE id=$1`, orderID, invoice.ID)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return s.checkoutLinkForOrder(ctx, businessID, req.ExternalOrderID)
}

func (s *MerchantIntegrationService) GetOrder(ctx context.Context, businessID, externalOrderID string) (*MerchantOrderResponse, error) {
	var result MerchantOrderResponse
	err := s.db.GetContext(ctx, &result, `SELECT o.external_order_id, o.invoice_id, i.status AS invoice_status
 FROM merchant_orders o JOIN invoices i ON i.id=o.invoice_id
 WHERE o.business_id=$1 AND o.external_order_id=$2`, businessID, externalOrderID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, utils.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *MerchantIntegrationService) checkoutLinkForOrder(ctx context.Context, businessID, externalOrderID string) (*MerchantOrderResponse, error) {
	result, err := s.GetOrder(ctx, businessID, externalOrderID)
	if err != nil || result.InvoiceStatus != "pending" {
		return result, err
	}
	link, err := s.checkout.CreateLink(ctx, result.InvoiceID, businessID)
	if err != nil {
		return nil, err
	}
	result.CheckoutURL = link.CheckoutURL
	return result, nil
}
