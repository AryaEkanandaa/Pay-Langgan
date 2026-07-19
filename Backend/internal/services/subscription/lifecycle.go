package subscription

import (
	"fmt"
	"time"

	"pay-langgan/internal/models"
	"pay-langgan/internal/utils"
)

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
