package subscription

import (
	"fmt"

	"pay-langgan/internal/models"
	"pay-langgan/internal/utils"
)

func (s *SubscriptionService) AddAddOn(id int, businessID string, userID int, addOnID, quantity int) (*models.SubscriptionDetailResponse, error) {
	if addOnID < 1 || quantity < 1 {
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

	return s.buildDetailResponse(sub, businessID)
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

	removed, err := s.subAddOnRepo.DeleteBySubscriptionIDAndAddOnID(tx, sub.ID, addOnID)
	if err != nil {
		return nil, fmt.Errorf("delete add-on link: %w", err)
	}
	if !removed {
		return nil, utils.ErrNotFound
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

	return s.buildDetailResponse(sub, businessID)
}
