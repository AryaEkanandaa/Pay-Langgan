package payment

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"pay-langgan/internal/models"
	"strings"
	"time"

	"pay-langgan/internal/utils"
)

type EmbedSite struct {
	ID        int64      `db:"id" json:"id"`
	Name      string     `db:"name" json:"-"`
	PublicID  string     `db:"public_id" json:"public_id"`
	Origin    string     `db:"origin" json:"-"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	RevokedAt *time.Time `db:"revoked_at" json:"revoked_at,omitempty"`
}

type CreateEmbedSiteRequest struct {
	Name   string `json:"name"`
	Origin string `json:"origin"`
}

func normalizeEmbedOrigin(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || u.User != nil || (u.Path != "" && u.Path != "/") || u.RawQuery != "" || u.Fragment != "" || u.ForceQuery || strings.Contains(raw, "*") ||
		(u.Scheme != "https" && !(u.Scheme == "http" && (u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1"))) {
		return "", fmt.Errorf("%w: website origin must be HTTPS (HTTP localhost is allowed for development), without a path", utils.ErrBadRequest)
	}
	return strings.ToLower(u.Scheme + "://" + u.Host), nil
}

func (s *MerchantIntegrationService) CreateEmbedSite(ctx context.Context, businessID string, request CreateEmbedSiteRequest) (*EmbedSite, error) {
	name := strings.TrimSpace(request.Name)
	if name == "" {
		name = "Integrasi bisnis"
	}
	origin := ""
	if strings.TrimSpace(request.Origin) != "" {
		var err error
		origin, err = normalizeEmbedOrigin(request.Origin)
		if err != nil {
			return nil, err
		}
	}
	if businessID == "" || len(name) > 100 || len(origin) > 255 {
		return nil, utils.ErrBadRequest
	}
	var existing EmbedSite
	err := s.db.GetContext(ctx, &existing, `SELECT id,name,public_id,origin,created_at,revoked_at
 FROM merchant_embed_sites WHERE business_id=$1 AND revoked_at IS NULL ORDER BY id DESC LIMIT 1`, businessID)
	if err == nil {
		if _, err = s.db.ExecContext(ctx, `UPDATE merchant_embed_sites SET revoked_at=(NOW() AT TIME ZONE 'UTC')
 WHERE business_id=$1 AND revoked_at IS NULL AND id<>$2`, businessID, existing.ID); err != nil {
			return nil, err
		}
		if strings.TrimSpace(request.Name) == "" && strings.TrimSpace(request.Origin) == "" {
			existing.Name = "Integrasi bisnis"
			existing.Origin = ""
			if _, err = s.db.ExecContext(ctx, `UPDATE merchant_embed_sites SET name=$2,origin=''
 WHERE id=$1`, existing.ID, existing.Name); err != nil {
				return nil, err
			}
		}
		return &existing, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	random, err := secureToken(18)
	if err != nil {
		return nil, err
	}
	site := &EmbedSite{Name: name, PublicID: "pl_site_" + random, Origin: origin}
	err = s.db.QueryRowContext(ctx, `INSERT INTO merchant_embed_sites(business_id,name,public_id,origin)
 VALUES ($1,$2,$3,$4) RETURNING id,created_at`, businessID, site.Name, site.PublicID, site.Origin).Scan(&site.ID, &site.CreatedAt)
	if err != nil {
		return nil, err
	}
	return site, nil
}

func (s *MerchantIntegrationService) ListEmbedSites(ctx context.Context, businessID string) ([]EmbedSite, error) {
	sites := []EmbedSite{}
	err := s.db.SelectContext(ctx, &sites, `SELECT id,name,public_id,origin,created_at,revoked_at
 FROM merchant_embed_sites WHERE business_id=$1 ORDER BY id DESC LIMIT 100`, businessID)
	if err != nil {
		return nil, err
	}
	return sites, nil
}

func (s *MerchantIntegrationService) RevokeEmbedSite(ctx context.Context, businessID string, siteID int64) (bool, error) {
	result, err := s.db.ExecContext(ctx, `UPDATE merchant_embed_sites SET revoked_at=(NOW() AT TIME ZONE 'UTC')
 WHERE id=$1 AND business_id=$2 AND revoked_at IS NULL`, siteID, businessID)
	if err != nil {
		return false, err
	}
	n, err := result.RowsAffected()
	return n == 1, err
}

type EmbedOrderRequest struct {
	RequestID    string          `json:"request_id,omitempty"`
	AddOns       []CheckoutAddOn `json:"add_ons,omitempty"`
	CouponCode   string          `json:"coupon_code,omitempty"`
	PlanPublicID string          `json:"plan_public_id"`
	PlanID       int             `json:"plan_id"` // Legacy buttons created before public plan IDs.
	Customer     struct {
		Name   string `json:"name"`
		Email  string `json:"email"`
		Mobile string `json:"mobile"`
	} `json:"customer"`
}

// Preview is public catalog data only. No customer, order or gateway credentials
// are created or exposed while a visitor reviews a package.
type EmbedPlanPreview struct {
	BusinessName string                           `db:"business_name" json:"business_name"`
	PlanName     string                           `db:"plan_name" json:"plan_name"`
	BillingCycle string                           `db:"billing_cycle" json:"billing_cycle"`
	Amount       float64                          `db:"amount" json:"amount"`
	TrialDays    int                              `json:"trial_days"`
	DueNow       float64                          `json:"due_now"`
	AddOns       []CheckoutAddOnOption            `json:"add_ons"`
	Items        []models.SubscriptionPreviewItem `json:"items"`
	CouponCode   string                           `json:"coupon_code,omitempty"`
}

func (s *MerchantIntegrationService) PreviewEmbedPlan(ctx context.Context, siteID, planID string) (*EmbedPlanPreview, error) {
	return s.QuoteEmbedPlan(ctx, siteID, planID, nil, "")
}

func (s *MerchantIntegrationService) QuoteEmbedPlan(ctx context.Context, siteID, planID string, options []CheckoutAddOn, coupon string) (*EmbedPlanPreview, error) {
	if !strings.HasPrefix(siteID, "pl_site_") || len(siteID) > 64 || len(planID) > 40 {
		return nil, utils.ErrNotFound
	}
	var preview EmbedPlanPreview
	err := s.db.GetContext(ctx, &preview, `SELECT b.name AS business_name, p.name AS plan_name, p.billing_cycle, p.price AS amount
 FROM merchant_embed_sites es JOIN businesses b ON b.id=es.business_id
 JOIN services sv ON sv.business_id=b.id JOIN products pr ON pr.service_id=sv.id JOIN plans p ON p.product_id=pr.id
 WHERE es.public_id=$1 AND es.revoked_at IS NULL AND b.status='active' AND b.deleted=FALSE AND pr.status='active'
 AND (p.public_id=$2 OR p.id::text=$2) AND p.billing_cycle IN ('monthly','yearly')`, siteID, planID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, utils.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	var owner struct {
		BusinessID string `db:"business_id"`
		ProductID  int    `db:"product_id"`
		TrialDays  int    `db:"trial_days"`
	}
	err = s.db.GetContext(ctx, &owner, `SELECT sv.business_id,p.product_id,p.trial_days FROM plans p JOIN products pr ON pr.id=p.product_id JOIN services sv ON sv.id=pr.service_id JOIN merchant_embed_sites es ON es.business_id=sv.business_id WHERE es.public_id=$1 AND (p.public_id=$2 OR p.id::text=$2)`, siteID, planID)
	if err != nil {
		return nil, err
	}
	plan := checkoutCatalogPlan{ProductID: owner.ProductID, Name: preview.PlanName, Price: preview.Amount, BillingCycle: preview.BillingCycle, TrialDays: owner.TrialDays}
	preview.Items, preview.Amount, _, err = quoteCheckout(ctx, s.db, owner.BusinessID, plan, options, strings.TrimSpace(coupon), false)
	if err != nil {
		return nil, err
	}
	preview.TrialDays = owner.TrialDays
	preview.DueNow = preview.Amount
	preview.CouponCode = strings.TrimSpace(coupon)
	if preview.TrialDays > 0 {
		preview.DueNow = 0
	}
	preview.AddOns = []CheckoutAddOnOption{}
	err = s.db.SelectContext(ctx, &preview.AddOns, `SELECT id,name,price FROM add_ons WHERE product_id=$1 AND billing_cycle=$2 AND visible_in_checkout=TRUE ORDER BY id`, owner.ProductID, preview.BillingCycle)
	if err != nil {
		return nil, err
	}
	return &preview, nil
}

// CreateEmbedOrder uses the business integration ID and its catalog. The public site ID
// is not a secret or merchant auth key; Origin is not used as an access-control boundary.
func (s *MerchantIntegrationService) CreateEmbedOrder(ctx context.Context, publicID, _ string, request EmbedOrderRequest) (*MerchantOrderResponse, error) {
	if !strings.HasPrefix(publicID, "pl_site_") || len(publicID) > 64 ||
		(request.PlanPublicID == "" && request.PlanID < 1) {
		return nil, utils.ErrNotFound
	}
	var site struct {
		BusinessID string `db:"business_id"`
	}
	err := s.db.GetContext(ctx, &site, `SELECT es.business_id FROM merchant_embed_sites es
 JOIN businesses b ON b.id=es.business_id
 WHERE es.public_id=$1 AND es.revoked_at IS NULL AND b.status='active' AND b.deleted=FALSE`, publicID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, utils.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	planID := request.PlanID
	if request.PlanPublicID != "" {
		if len(request.PlanPublicID) != 32 || !strings.HasPrefix(request.PlanPublicID, "pl_plan_") {
			return nil, utils.ErrNotFound
		}
		err = s.db.GetContext(ctx, &planID, `SELECT p.id FROM plans p
 JOIN products pr ON pr.id=p.product_id JOIN services sv ON sv.id=pr.service_id
 WHERE p.public_id=$1 AND sv.business_id=$2`, request.PlanPublicID, site.BusinessID)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, utils.ErrNotFound
		}
		if err != nil {
			return nil, err
		}
	}
	reference, err := secureToken(18)
	if err != nil {
		return nil, err
	}
	order := MerchantOrderRequest{ExternalOrderID: "embed_" + reference, PlanID: planID}
	if request.RequestID != "" {
		if len(request.RequestID) > 64 || strings.ContainsAny(request.RequestID, "\r\n\t ") {
			return nil, utils.ErrBadRequest
		}
		order.ExternalOrderID = "embed_" + merchantKeyHash(publicID+":"+request.RequestID)
	}
	order.AddOns = request.AddOns
	order.CouponCode = request.CouponCode
	order.Customer.Name = request.Customer.Name
	order.Customer.Email = request.Customer.Email
	order.Customer.Mobile = request.Customer.Mobile
	return s.CreateOrder(ctx, site.BusinessID, order)
}

func (s *MerchantIntegrationService) AllowEmbedRequest(ctx context.Context, key string, limit int) (bool, error) {
	return s.checkout.AllowCheckoutRequest(ctx, "embed:"+key, limit)
}
