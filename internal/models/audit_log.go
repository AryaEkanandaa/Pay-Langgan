package models

import (
	"time"
)

type AuditLog struct {
	ID         int       `db:"id" json:"id"`
	BusinessID string    `db:"business_id" json:"business_id"`
	UserID     *int      `db:"user_id" json:"user_id"`
	Action     string    `db:"action" json:"action"`
	EntityType string    `db:"entity_type" json:"entity_type"`
	EntityID   *string   `db:"entity_id" json:"entity_id"`
	OldValue   JSONMap   `db:"old_value" json:"old_value"`
	NewValue   JSONMap   `db:"new_value" json:"new_value"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}
