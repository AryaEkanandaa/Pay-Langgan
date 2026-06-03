package repositories

import (
	"fmt"

	"pay-langgan/internal/database"
	"pay-langgan/internal/models"

	"github.com/jmoiron/sqlx"
)

type AuditLogRepository struct {
	db *database.DB
}

func NewAuditLogRepository(db *database.DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

func (r *AuditLogRepository) Create(tx *sqlx.Tx, log *models.AuditLog) error {
	if log.OldValue == nil {
		log.OldValue = make(models.JSONMap)
	}
	if log.NewValue == nil {
		log.NewValue = make(models.JSONMap)
	}
	query := `INSERT INTO audit_logs (business_id, user_id, action, entity_type, entity_id, old_value, new_value, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())`
	_, err := tx.Exec(query, log.BusinessID, log.UserID, log.Action, log.EntityType, log.EntityID, log.OldValue, log.NewValue)
	if err != nil {
		return fmt.Errorf("create audit log: %w", err)
	}
	return nil
}
