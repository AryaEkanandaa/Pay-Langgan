package subscription

import (
	"fmt"
	"time"

	"pay-langgan/internal/models"
	"pay-langgan/internal/utils"
)

func (s *SubscriptionService) ApplyCoupon(id int, businessID string, userID int, couponCode string) (*models.SubscriptionDetailResponse, error) {
	sub, err := s.subRepo.FindByIDAndBusinessID(id, businessID)
	if err != nil {
		return nil, fmt.Errorf("find subscription: %w", err)
	}
	if sub == nil {
		return nil, utils.ErrNotFound
	}

	coup, err := s.couponRepo.FindByCode(couponCode)
	if err != nil {
		return nil, fmt.Errorf("find coupon: %w", err)
	}
	if coup == nil {
		return nil, fmt.Errorf("coupon %s not found", couponCode)
	}

	if coup.ExpiresAt != nil && coup.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("coupon has expired")
	}
	if coup.MaxUsage != nil && coup.UsedCount >= *coup.MaxUsage {
		return nil, fmt.Errorf("coupon usage limit exceeded")
	}

	exists, err := s.subCpnRepo.ExistsBySubscriptionIDAndCouponID(sub.ID, coup.ID)
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
		CouponID:       coup.ID,
	}
	if err := s.subCpnRepo.Apply(tx, subCpn); err != nil {
		return nil, fmt.Errorf("apply coupon: %w", err)
	}

	coup.UsedCount++
	if err := s.couponRepo.UpdateTx(tx, coup); err != nil {
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

	coup, err := s.couponRepo.FindByID(couponID)
	if err != nil {
		return nil, fmt.Errorf("find coupon: %w", err)
	}
	if coup == nil {
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

	if coup.UsedCount > 0 {
		coup.UsedCount--
		if err := s.couponRepo.UpdateTx(tx, coup); err != nil {
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
