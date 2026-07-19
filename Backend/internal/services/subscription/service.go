package subscription

import (
	"fmt"
	"time"

	"pay-langgan/internal/database"
	"pay-langgan/internal/models"
	"pay-langgan/internal/repositories/audit"
	"pay-langgan/internal/repositories/catalog"
	"pay-langgan/internal/repositories/coupon"
	"pay-langgan/internal/repositories/customer"
	"pay-langgan/internal/repositories/subscription"
	"pay-langgan/internal/utils"
)

type SubscriptionService struct {
	db           *database.DB
	subRepo      *subscription.SubscriptionRepository
	subAddOnRepo *subscription.SubscriptionAddOnRepository
	subCpnRepo   *subscription.SubscriptionCouponRepository
	auditLogRepo *audit.AuditLogRepository
	planRepo     *catalog.PlanRepository
	productRepo  *catalog.ProductRepository
	serviceRepo  *catalog.ServiceRepository
	addOnRepo    *catalog.AddOnRepository
	couponRepo   *coupon.CouponRepository
	customerRepo *customer.CustomerRepository
}

func NewSubscriptionService(
	db *database.DB,
	subRepo *subscription.SubscriptionRepository,
	subAddOnRepo *subscription.SubscriptionAddOnRepository,
	subCpnRepo *subscription.SubscriptionCouponRepository,
	auditLogRepo *audit.AuditLogRepository,
	planRepo *catalog.PlanRepository,
	productRepo *catalog.ProductRepository,
	serviceRepo *catalog.ServiceRepository,
	addOnRepo *catalog.AddOnRepository,
	couponRepo *coupon.CouponRepository,
	customerRepo *customer.CustomerRepository,
) *SubscriptionService {
	return &SubscriptionService{
		db:           db,
		subRepo:      subRepo,
		subAddOnRepo: subAddOnRepo,
		subCpnRepo:   subCpnRepo,
		auditLogRepo: auditLogRepo,
		planRepo:     planRepo,
		productRepo:  productRepo,
		serviceRepo:  serviceRepo,
		addOnRepo:    addOnRepo,
		couponRepo:   couponRepo,
		customerRepo: customerRepo,
	}
}

func (s *SubscriptionService) List(businessID string, page, limit int, status, search string) ([]models.Subscription, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	return s.subRepo.FindAllByBusinessID(businessID, page, limit, status, search)
}

func (s *SubscriptionService) GetDetail(id int, businessID string) (*models.SubscriptionDetailResponse, error) {
	sub, err := s.subRepo.FindByIDAndBusinessID(id, businessID)
	if err != nil {
		return nil, err
	}
	if sub == nil {
		return nil, utils.ErrNotFound
	}

	return s.buildDetailResponse(sub)
}

func (s *SubscriptionService) Create(businessID string, userID int, req models.CreateSubscriptionRequest) (*models.SubscriptionDetailResponse, error) {
	if req.CustomerID == 0 || req.PlanID == 0 {
		return nil, utils.ErrBadRequest
	}

	cust, err := s.customerRepo.FindByIDAndBusinessID(req.CustomerID, businessID)
	if err != nil {
		return nil, err
	}
	if cust == nil {
		return nil, utils.ErrNotFound
	}

	plan, err := s.planRepo.FindByIDAndBusinessID(req.PlanID, businessID)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		return nil, utils.ErrNotFound
	}

	now := time.Now()
	sub := &models.Subscription{
		CustomerID: req.CustomerID,
		PlanID:     req.PlanID,
		StartDate:  now,
		Meta:       req.Meta,
	}

	if plan.TrialDays > 0 {
		sub.Status = "trial"
		trialEnd := now.AddDate(0, 0, plan.TrialDays)
		sub.TrialEndsAt = &trialEnd
		sub.NextBillingDate = &trialEnd
	} else {
		sub.Status = "active"
		sub.NextBillingDate = &now
	}

	tx, err := s.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if err := s.subRepo.Create(tx, sub); err != nil {
		return nil, fmt.Errorf("create subscription: %w", err)
	}

	for _, a := range req.AddOns {
		if a.Quantity < 1 {
			continue
		}
		addOn, err := s.addOnRepo.FindByIDAndBusinessID(a.AddOnID, businessID)
		if err != nil {
			return nil, fmt.Errorf("find add-on %d: %w", a.AddOnID, err)
		}
		if addOn == nil {
			return nil, fmt.Errorf("add-on %d not found", a.AddOnID)
		}

		item := &models.SubscriptionAddOn{
			SubscriptionID: sub.ID,
			AddOnID:        a.AddOnID,
			Quantity:       a.Quantity,
		}
		if err := s.subAddOnRepo.Upsert(tx, item); err != nil {
			return nil, fmt.Errorf("upsert add-on %d: %w", a.AddOnID, err)
		}
	}

	if req.CouponCode != "" {
		coup, err := s.couponRepo.FindByCode(req.CouponCode)
		if err != nil {
			return nil, fmt.Errorf("find coupon: %w", err)
		}
		if coup == nil {
			return nil, fmt.Errorf("coupon %s not found", req.CouponCode)
		}
		if coup.ExpiresAt != nil && coup.ExpiresAt.Before(time.Now()) {
			return nil, fmt.Errorf("coupon has expired")
		}
		if coup.MaxUsage != nil && coup.UsedCount >= *coup.MaxUsage {
			return nil, fmt.Errorf("coupon usage limit exceeded")
		}

		subCpn := &models.SubscriptionCoupon{
			SubscriptionID: sub.ID,
			CouponID:       coup.ID,
		}
		if err := s.subCpnRepo.Apply(tx, subCpn); err != nil {
			return nil, fmt.Errorf("apply coupon: %w", err)
		}

		coup.UsedCount++
		if err := s.couponRepo.UpdateTx(tx, coup); err != nil {
			return nil, fmt.Errorf("update coupon usage: %w", err)
		}
	}

	entityID := fmt.Sprintf("%d", sub.ID)
	auditLog := &models.AuditLog{
		BusinessID: businessID,
		UserID:     &userID,
		Action:     "create_subscription",
		EntityType: "subscription",
		EntityID:   &entityID,
	}
	if err := s.auditLogRepo.Create(tx, auditLog); err != nil {
		return nil, fmt.Errorf("create audit log: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return s.buildDetailResponse(sub)
}
