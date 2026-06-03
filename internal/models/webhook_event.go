package models

import (
	"time"
)

type WebhookEvent struct {
	ID          int            `db:"id" json:"id"`
	Provider    string         `db:"provider" json:"provider"`
	EventType   string         `db:"event_type" json:"event_type"`
	ReferenceID *string        `db:"reference_id" json:"reference_id"`
	Payload     map[string]any `db:"payload" json:"payload"`
	Processed   bool           `db:"processed" json:"processed"`
	ReceivedAt  time.Time      `db:"received_at" json:"received_at"`
	ProcessedAt *time.Time     `db:"processed_at" json:"processed_at"`
	CreatedAt   time.Time      `db:"created_at" json:"created_at"`
}
