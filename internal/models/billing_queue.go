package models

import (
	"time"
)

type BillingQueue struct {
	ID             int            `db:"id" json:"id"`
	SubscriptionID int            `db:"subscription_id" json:"subscription_id"`
	ScheduledAt    time.Time      `db:"scheduled_at" json:"scheduled_at"`
	Processed      bool           `db:"processed" json:"processed"`
	RetryCount     int            `db:"retry_count" json:"retry_count"`
	LastError      *string        `db:"last_error" json:"last_error"`
	ProcessedAt    *time.Time     `db:"processed_at" json:"processed_at"`
	CreatedAt      time.Time      `db:"created_at" json:"created_at"`
}
