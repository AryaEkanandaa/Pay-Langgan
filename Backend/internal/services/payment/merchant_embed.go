package payment

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
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
	PlanPublicID string `json:"plan_public_id"`
	PlanID       int    `json:"plan_id"` // Legacy buttons created before public plan IDs.
	Customer     struct {
		Name   string `json:"name"`
		Email  string `json:"email"`
		Mobile string `json:"mobile"`
	} `json:"customer"`
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
	order.Customer.Name = request.Customer.Name
	order.Customer.Email = request.Customer.Email
	order.Customer.Mobile = request.Customer.Mobile
	return s.CreateOrder(ctx, site.BusinessID, order)
}

func (s *MerchantIntegrationService) AllowEmbedRequest(ctx context.Context, key string, limit int) (bool, error) {
	return s.checkout.AllowCheckoutRequest(ctx, "embed:"+key, limit)
}
