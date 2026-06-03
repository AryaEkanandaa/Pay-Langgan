package services

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"pay-langgan/internal/database"
	"pay-langgan/internal/models"
	"pay-langgan/internal/repositories"
	"pay-langgan/internal/utils"
)

type SubscriptionService struct {
	db           *database.DB
	subRepo      *repositories.SubscriptionRepository
	subAddOnRepo *repositories.SubscriptionAddOnRepository
	subCpnRepo   *repositories.SubscriptionCouponRepository
	auditLogRepo *repositories.AuditLogRepository
	planRepo     *repositories.PlanRepository
	productRepo  *repositories.ProductRepository
	serviceRepo  *repositories.ServiceRepository
	addOnRepo    *repositories.AddOnRepository
	couponRepo   *repositories.CouponRepository
	customerRepo *repositories.CustomerRepository
}

func NewSubscriptionService(
	db *database.DB,
	subRepo *repositories.SubscriptionRepository,
	subAddOnRepo *repositories.SubscriptionAddOnRepository,
	subCpnRepo *repositories.SubscriptionCouponRepository,
	auditLogRepo *repositories.AuditLogRepository,
	planRepo *repositories.PlanRepository,
	productRepo *repositories.ProductRepository,
	serviceRepo *repositories.ServiceRepository,
	addOnRepo *repositories.AddOnRepository,
	couponRepo *repositories.CouponRepository,
	customerRepo *repositories.CustomerRepository,
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

	customer, err := s.customerRepo.FindByIDAndBusinessID(req.CustomerID, businessID)
	if err != nil {
		return nil, err
	}
	if customer == nil {
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
		coupon, err := s.couponRepo.FindByCode(req.CouponCode)
		if err != nil {
			return nil, fmt.Errorf("find coupon: %w", err)
		}
		if coupon == nil {
			return nil, fmt.Errorf("coupon %s not found", req.CouponCode)
		}
		if coupon.ExpiresAt != nil && coupon.ExpiresAt.Before(time.Now()) {
			return nil, fmt.Errorf("coupon has expired")
		}
		if coupon.MaxUsage != nil && coupon.UsedCount >= *coupon.MaxUsage {
			return nil, fmt.Errorf("coupon usage limit exceeded")
		}

		subCpn := &models.SubscriptionCoupon{
			SubscriptionID: sub.ID,
			CouponID:       coupon.ID,
		}
		if err := s.subCpnRepo.Apply(tx, subCpn); err != nil {
			return nil, fmt.Errorf("apply coupon: %w", err)
		}

		coupon.UsedCount++
		if err := s.couponRepo.UpdateTx(tx, coupon); err != nil {
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

func (s *SubscriptionService) Cancel(id int, businessID string, userID int, reason string) (*models.SubscriptionDetailResponse, error) {
	sub, err := s.subRepo.FindByIDAndBusinessID(id, businessID)
	if err != nil {
		return nil, fmt.Errorf("find subscription: %w", err)
	}
	if sub == nil {
		return nil, utils.ErrNotFound
	}

	if sub.Status != "trial" && sub.Status != "active" && sub.Status != "paused" {
		return nil, fmt.Errorf("cannot cancel subscription with status %s", sub.Status)
	}

	now := time.Now()
	sub.Status = "cancelled"
	sub.EndDate = &now
	if sub.Meta == nil {
		sub.Meta = make(models.JSONMap)
	}
	if reason != "" {
		sub.Meta["cancel_reason"] = reason
	}

	tx, err := s.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if err := s.subRepo.UpdateStatus(tx, sub); err != nil {
		return nil, fmt.Errorf("update status: %w", err)
	}

	entityID := fmt.Sprintf("%d", sub.ID)
	auditLog := &models.AuditLog{
		BusinessID: businessID,
		UserID:     &userID,
		Action:     "cancel_subscription",
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

func (s *SubscriptionService) Pause(id int, businessID string, userID int, reason string) (*models.SubscriptionDetailResponse, error) {
	sub, err := s.subRepo.FindByIDAndBusinessID(id, businessID)
	if err != nil {
		return nil, fmt.Errorf("find subscription: %w", err)
	}
	if sub == nil {
		return nil, utils.ErrNotFound
	}

	if sub.Status != "active" {
		return nil, fmt.Errorf("cannot pause subscription with status %s", sub.Status)
	}

	sub.Status = "paused"
	if sub.Meta == nil {
		sub.Meta = make(models.JSONMap)
	}
	if reason != "" {
		sub.Meta["pause_reason"] = reason
	}

	tx, err := s.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if err := s.subRepo.UpdateStatus(tx, sub); err != nil {
		return nil, fmt.Errorf("update status: %w", err)
	}

	entityID := fmt.Sprintf("%d", sub.ID)
	auditLog := &models.AuditLog{
		BusinessID: businessID,
		UserID:     &userID,
		Action:     "pause_subscription",
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

func (s *SubscriptionService) Resume(id int, businessID string, userID int) (*models.SubscriptionDetailResponse, error) {
	sub, err := s.subRepo.FindByIDAndBusinessID(id, businessID)
	if err != nil {
		return nil, fmt.Errorf("find subscription: %w", err)
	}
	if sub == nil {
		return nil, utils.ErrNotFound
	}

	if sub.Status != "paused" {
		return nil, fmt.Errorf("cannot resume subscription with status %s", sub.Status)
	}

	now := time.Now()
	sub.Status = "active"
	if sub.NextBillingDate == nil || sub.NextBillingDate.Before(now) {
		sub.NextBillingDate = &now
	}

	tx, err := s.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if err := s.subRepo.UpdateStatus(tx, sub); err != nil {
		return nil, fmt.Errorf("update status: %w", err)
	}

	entityID := fmt.Sprintf("%d", sub.ID)
	auditLog := &models.AuditLog{
		BusinessID: businessID,
		UserID:     &userID,
		Action:     "resume_subscription",
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

func (s *SubscriptionService) AddAddOn(id int, businessID string, userID int, addOnID, quantity int) (*models.SubscriptionDetailResponse, error) {
	if quantity < 1 {
		return nil, utils.ErrBadRequest
	}

	sub, err := s.subRepo.FindByIDAndBusinessID(id, businessID)
	if err != nil {
		return nil, fmt.Errorf("find subscription: %w", err)
	}
	if sub == nil {
		return nil, utils.ErrNotFound
	}

	addOn, err := s.addOnRepo.FindByIDAndBusinessID(addOnID, businessID)
	if err != nil {
		return nil, fmt.Errorf("find add-on: %w", err)
	}
	if addOn == nil {
		return nil, utils.ErrNotFound
	}

	tx, err := s.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	item := &models.SubscriptionAddOn{
		SubscriptionID: sub.ID,
		AddOnID:        addOnID,
		Quantity:       quantity,
	}
	if err := s.subAddOnRepo.Upsert(tx, item); err != nil {
		return nil, fmt.Errorf("upsert add-on: %w", err)
	}

	entityID := fmt.Sprintf("%d", sub.ID)
	auditLog := &models.AuditLog{
		BusinessID: businessID,
		UserID:     &userID,
		Action:     "add_subscription_add_on",
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

func (s *SubscriptionService) RemoveAddOn(id int, businessID string, userID int, addOnID int) (*models.SubscriptionDetailResponse, error) {
	sub, err := s.subRepo.FindByIDAndBusinessID(id, businessID)
	if err != nil {
		return nil, fmt.Errorf("find subscription: %w", err)
	}
	if sub == nil {
		return nil, utils.ErrNotFound
	}

	tx, err := s.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if err := s.subAddOnRepo.DeleteBySubscriptionIDAndAddOnID(tx, sub.ID, addOnID); err != nil {
		return nil, fmt.Errorf("delete add-on link: %w", err)
	}

	entityID := fmt.Sprintf("%d", sub.ID)
	auditLog := &models.AuditLog{
		BusinessID: businessID,
		UserID:     &userID,
		Action:     "remove_subscription_add_on",
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

func (s *SubscriptionService) ApplyCoupon(id int, businessID string, userID int, couponCode string) (*models.SubscriptionDetailResponse, error) {
	sub, err := s.subRepo.FindByIDAndBusinessID(id, businessID)
	if err != nil {
		return nil, fmt.Errorf("find subscription: %w", err)
	}
	if sub == nil {
		return nil, utils.ErrNotFound
	}

	coupon, err := s.couponRepo.FindByCode(couponCode)
	if err != nil {
		return nil, fmt.Errorf("find coupon: %w", err)
	}
	if coupon == nil {
		return nil, fmt.Errorf("coupon %s not found", couponCode)
	}

	if coupon.ExpiresAt != nil && coupon.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("coupon has expired")
	}
	if coupon.MaxUsage != nil && coupon.UsedCount >= *coupon.MaxUsage {
		return nil, fmt.Errorf("coupon usage limit exceeded")
	}

	exists, err := s.subCpnRepo.ExistsBySubscriptionIDAndCouponID(sub.ID, coupon.ID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, utils.ErrConflict
	}

	tx, err := s.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	subCpn := &models.SubscriptionCoupon{
		SubscriptionID: sub.ID,
		CouponID:       coupon.ID,
	}
	if err := s.subCpnRepo.Apply(tx, subCpn); err != nil {
		return nil, fmt.Errorf("apply coupon: %w", err)
	}

	coupon.UsedCount++
	if err := s.couponRepo.UpdateTx(tx, coupon); err != nil {
		return nil, fmt.Errorf("update coupon usage: %w", err)
	}

	entityID := fmt.Sprintf("%d", sub.ID)
	auditLog := &models.AuditLog{
		BusinessID: businessID,
		UserID:     &userID,
		Action:     "apply_subscription_coupon",
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

func (s *SubscriptionService) RemoveCoupon(id int, businessID string, userID int, couponID int) (*models.SubscriptionDetailResponse, error) {
	sub, err := s.subRepo.FindByIDAndBusinessID(id, businessID)
	if err != nil {
		return nil, fmt.Errorf("find subscription: %w", err)
	}
	if sub == nil {
		return nil, utils.ErrNotFound
	}

	coupon, err := s.couponRepo.FindByID(couponID)
	if err != nil {
		return nil, fmt.Errorf("find coupon: %w", err)
	}
	if coupon == nil {
		return nil, utils.ErrNotFound
	}

	tx, err := s.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if err := s.subCpnRepo.RemoveBySubscriptionIDAndCouponID(tx, sub.ID, couponID); err != nil {
		return nil, fmt.Errorf("remove coupon link: %w", err)
	}

	if coupon.UsedCount > 0 {
		coupon.UsedCount--
		if err := s.couponRepo.UpdateTx(tx, coupon); err != nil {
			return nil, fmt.Errorf("update coupon usage: %w", err)
		}
	}

	entityID := fmt.Sprintf("%d", sub.ID)
	auditLog := &models.AuditLog{
		BusinessID: businessID,
		UserID:     &userID,
		Action:     "remove_subscription_coupon",
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

func (s *SubscriptionService) buildDetailResponse(sub *models.Subscription) (*models.SubscriptionDetailResponse, error) {
	resp := &models.SubscriptionDetailResponse{
		ID:        sub.ID,
		CustomerID: sub.CustomerID,
		PlanID:    sub.PlanID,
		Status:    sub.Status,
		StartDate: sub.StartDate.Format(time.RFC3339),
		Meta:      sub.Meta,
		CreatedAt: sub.CreatedAt.Format(time.RFC3339),
	}

	if sub.NextBillingDate != nil {
		v := sub.NextBillingDate.Format(time.RFC3339)
		resp.NextBillingDate = &v
	}
	if sub.EndDate != nil {
		v := sub.EndDate.Format(time.RFC3339)
		resp.EndDate = &v
	}
	if sub.TrialEndsAt != nil {
		v := sub.TrialEndsAt.Format(time.RFC3339)
		resp.TrialEndsAt = &v
	}

	customer, err := s.customerRepo.FindByID(sub.CustomerID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if customer != nil {
		resp.Customer = &models.SubscriptionCustomerInfo{
			ID:    customer.ID,
			Name:  customer.Name,
			Email: customer.Email,
		}
	}

	plan, err := s.planRepo.FindByID(sub.PlanID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if plan != nil {
		resp.Plan = &models.SubscriptionPlanInfo{
			ID:           plan.ID,
			Name:         plan.Name,
			Price:        plan.Price,
			BillingCycle: plan.BillingCycle,
			TrialDays:    plan.TrialDays,
		}

		product, err := s.productRepo.FindByID(plan.ProductID)
		if err == nil && product != nil {
			resp.Product = &models.SubscriptionProductInfo{
				ID:   product.ID,
				Name: product.Name,
			}

			service, _ := s.serviceRepo.FindByID(product.ServiceID)
			if service != nil {
				resp.Service = &models.SubscriptionServiceInfo{
					ID:   service.ID,
					Name: service.Name,
				}
			}
		}
	}

	addOnItems, err := s.subAddOnRepo.FindBySubscriptionID(sub.ID)
	if err != nil {
		return nil, err
	}
	for _, a := range addOnItems {
		addOnResp := models.SubscriptionAddOnResponse{
			ID:       a.ID,
			AddOnID:  a.AddOnID,
			Quantity: a.Quantity,
		}
		addOn, _ := s.addOnRepo.FindByID(a.AddOnID)
		if addOn != nil {
			addOnResp.Name = addOn.Name
			addOnResp.Price = addOn.Price
		}
		resp.AddOns = append(resp.AddOns, addOnResp)
	}

	couponItems, err := s.subCpnRepo.FindBySubscriptionID(sub.ID)
	if err != nil {
		return nil, err
	}
	for _, sc := range couponItems {
		coupResp := models.SubscriptionCouponResponse{
			ID:       sc.ID,
			CouponID: sc.CouponID,
		}
		coupon, _ := s.couponRepo.FindByID(sc.CouponID)
		if coupon != nil {
			coupResp.Code = coupon.Code
			coupResp.DiscountType = coupon.DiscountType
			coupResp.DiscountValue = coupon.DiscountValue
		}
		resp.Coupons = append(resp.Coupons, coupResp)
	}

	return resp, nil
}
