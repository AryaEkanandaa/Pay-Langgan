package payment

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
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
	ExternalOrderID string          `json:"external_order_id"`
	PlanID          int             `json:"plan_id"`
	AddOns          []CheckoutAddOn `json:"add_ons,omitempty"`
	CouponCode      string          `json:"coupon_code,omitempty"`
	Customer        struct {
		Name   string `json:"name"`
		Email  string `json:"email"`
		Mobile string `json:"mobile"`
	} `json:"customer"`
}

type MerchantOrderResponse struct {
	ExternalOrderID    string     `db:"external_order_id" json:"external_order_id"`
	InvoiceID          int        `db:"invoice_id" json:"invoice_id"`
	InvoiceStatus      string     `db:"invoice_status" json:"invoice_status"`
	CheckoutURL        string     `json:"checkout_url,omitempty"`
	SubscriptionID     int        `db:"subscription_id" json:"subscription_id"`
	SubscriptionStatus string     `db:"subscription_status" json:"subscription_status"`
	TrialEndsAt        *time.Time `db:"trial_ends_at" json:"trial_ends_at,omitempty"`
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
	req.Customer.Email = strings.ToLower(strings.TrimSpace(req.Customer.Email))
	req.Customer.Mobile = strings.TrimSpace(req.Customer.Mobile)
	req.CouponCode = strings.TrimSpace(req.CouponCode)
	if err := validateCheckoutOptions(req.AddOns, req.CouponCode); err != nil {
		return req, err
	}
	req.AddOns = append([]CheckoutAddOn{}, req.AddOns...)
	sort.Slice(req.AddOns, func(i, j int) bool { return req.AddOns[i].AddOnID < req.AddOns[j].AddOnID })
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
	var plan checkoutCatalogPlan
	err = tx.GetContext(ctx, &plan, `SELECT p.product_id,p.name,p.price,p.billing_cycle,p.trial_days FROM plans p
 JOIN products pr ON pr.id=p.product_id JOIN services sv ON sv.id=pr.service_id
 WHERE p.id=$1 AND sv.business_id=$2 AND pr.status='active' FOR SHARE OF p,pr`, req.PlanID, businessID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, utils.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if plan.TrialDays < 0 || plan.Price < 5000 || plan.Price > 200000000 || (plan.BillingCycle != "monthly" && plan.BillingCycle != "yearly") {
		return nil, fmt.Errorf("%w: this plan is not supported by card checkout", utils.ErrBadRequest)
	}
	fingerprintData, _ := json.Marshal(struct {
		Plan   int
		Email  string
		AddOns []CheckoutAddOn
		Coupon string
	}{req.PlanID, strings.ToLower(req.Customer.Email), req.AddOns, req.CouponCode})
	fingerprint := merchantKeyHash(string(fingerprintData))
	var orderID int64
	err = tx.QueryRowContext(ctx, `INSERT INTO merchant_orders (business_id, external_order_id, plan_id, customer_email,request_hash)
 VALUES ($1,$2,$3,$4,$5) ON CONFLICT (business_id, external_order_id) DO NOTHING RETURNING id`, businessID, req.ExternalOrderID, req.PlanID, req.Customer.Email, fingerprint).Scan(&orderID)
	if errors.Is(err, sql.ErrNoRows) {
		var existing struct {
			PlanID        int     `db:"plan_id"`
			CustomerEmail string  `db:"customer_email"`
			RequestHash   *string `db:"request_hash"`
		}
		if err := tx.GetContext(ctx, &existing, `SELECT plan_id, customer_email,request_hash FROM merchant_orders
 WHERE business_id=$1 AND external_order_id=$2`, businessID, req.ExternalOrderID); err != nil {
			return nil, err
		}
		if existing.PlanID != req.PlanID || !strings.EqualFold(existing.CustomerEmail, req.Customer.Email) || (existing.RequestHash != nil && *existing.RequestHash != fingerprint) || (existing.RequestHash == nil && (len(req.AddOns) > 0 || req.CouponCode != "")) {
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
 WHERE c.business_id=$1 AND lower(btrim(c.email))=lower($2)
 ORDER BY EXISTS (SELECT 1 FROM payment_methods pm WHERE pm.customer_id=c.id AND pm.is_default=true) DESC, c.id
 LIMIT 1`, businessID, req.Customer.Email).Scan(&customerID)
	if errors.Is(err, sql.ErrNoRows) {
		err = tx.QueryRowContext(ctx, `INSERT INTO customers (business_id,name,email,contact,meta)
 VALUES ($1,$2,$3,$4,'{}'::jsonb) RETURNING id`, businessID, req.Customer.Name, req.Customer.Email, req.Customer.Mobile).Scan(&customerID)
	}
	if err != nil {
		return nil, err
	}
	// Recheck and lock coupon quota only after the idempotent retry branch.
	if err = s.checkout.releaseExpiredTrialCards(ctx, tx, 100); err != nil {
		return nil, err
	}
	quoteItems, quoteAmount, couponID, err := quoteCheckout(ctx, tx, businessID, plan, req.AddOns, req.CouponCode, true)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	anchor := now
	status := "pending"
	var trialEnd *time.Time
	if plan.TrialDays > 0 {
		var used bool
		if err = tx.GetContext(ctx, &used, `SELECT EXISTS(SELECT 1 FROM subscriptions s WHERE customer_id=$1 AND plan_id=$2 AND (trial_ends_at IS NOT NULL OR EXISTS (SELECT 1 FROM trial_card_sessions t WHERE t.subscription_id=s.id AND t.status NOT IN ('expired','cancelled') AND t.expires_at>NOW())))`, customerID, req.PlanID); err != nil {
			return nil, err
		}
		if used {
			return nil, fmt.Errorf("%w: masa trial untuk pelanggan dan paket ini sudah pernah digunakan", utils.ErrConflict)
		}
		anchor = now.AddDate(0, 0, plan.TrialDays)
		trialEnd = &anchor
		// Trial starts only after card enrollment. No trial clock or email yet.
	}
	next, err := billingsvc.NextPeriod(anchor, anchor, plan.BillingCycle)
	if err != nil {
		return nil, err
	}
	if trialEnd != nil {
		next = anchor
	}
	var subscriptionID int
	err = tx.QueryRowContext(ctx, `INSERT INTO subscriptions
 (customer_id,plan_id,status,start_date,billing_anchor,billing_cycle,next_billing_date,trial_ends_at,meta)
 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,'{}'::jsonb) RETURNING id`, customerID, req.PlanID, status, now, anchor, plan.BillingCycle, next, trialEnd).Scan(&subscriptionID)
	if err != nil {
		return nil, err
	}
	for _, item := range req.AddOns {
		if _, err = tx.ExecContext(ctx, `INSERT INTO subscription_add_ons(subscription_id,add_on_id,quantity) VALUES($1,$2,$3)`, subscriptionID, item.AddOnID, item.Quantity); err != nil {
			return nil, err
		}
	}
	if couponID > 0 {
		if _, err = tx.ExecContext(ctx, `INSERT INTO subscription_coupons(subscription_id,coupon_id) VALUES($1,$2)`, subscriptionID, couponID); err != nil {
			return nil, err
		}
		if _, err = tx.ExecContext(ctx, `UPDATE coupons SET used_count=used_count+1 WHERE id=$1`, couponID); err != nil {
			return nil, err
		}
	}
	var invoiceID *int
	if plan.TrialDays == 0 {
		invoice, issueErr := billingsvc.IssueInvoiceContext(ctx, tx, s.invoices, subscriptionID, now, next, plan.BillingCycle, true)
		if issueErr != nil {
			return nil, issueErr
		}
		invoiceID = &invoice.ID
	}
	if plan.TrialDays > 0 {
		if _, err = tx.ExecContext(ctx, `UPDATE subscriptions SET trial_ends_at=NULL WHERE id=$1`, subscriptionID); err != nil {
			return nil, err
		}
		if err = s.checkout.prepareTrialCard(ctx, tx, subscriptionID, businessID, plan.TrialDays, quoteAmount, quoteItems); err != nil {
			return nil, err
		}
	}
	if _, err = tx.ExecContext(ctx, `UPDATE merchant_orders SET invoice_id=$2,subscription_id=$3 WHERE id=$1`, orderID, invoiceID, subscriptionID); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return s.checkoutLinkForOrder(ctx, businessID, req.ExternalOrderID)
}

func (s *MerchantIntegrationService) GetOrder(ctx context.Context, businessID, externalOrderID string) (*MerchantOrderResponse, error) {
	var result MerchantOrderResponse
	err := s.db.GetContext(ctx, &result, `SELECT o.external_order_id, COALESCE(o.invoice_id,0) AS invoice_id, COALESCE(i.status,'not_issued') AS invoice_status,
 COALESCE(o.subscription_id,i.subscription_id,0) AS subscription_id,s.status AS subscription_status,s.trial_ends_at
 FROM merchant_orders o LEFT JOIN invoices i ON i.id=o.invoice_id JOIN subscriptions s ON s.id=COALESCE(o.subscription_id,i.subscription_id)
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
	if err == nil && result.InvoiceStatus == "not_issued" && result.SubscriptionStatus == "pending" {
		var token string
		if err = s.db.GetContext(ctx, &token, `SELECT public_token FROM trial_card_sessions WHERE subscription_id=$1`, result.SubscriptionID); err != nil {
			return nil, err
		}
		result.CheckoutURL = s.checkout.publicBaseURL + "/checkout/" + token
		return result, nil
	}
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
