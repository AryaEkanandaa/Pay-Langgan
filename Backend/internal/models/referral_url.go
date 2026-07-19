package models

import (
	"time"
)

type ReferralURL struct {
	ID         int       `db:"id" json:"id"`
	BusinessID string    `db:"business_id" json:"business_id"`
	Code       string    `db:"code" json:"code"`
	TargetURL  string    `db:"target_url" json:"target_url"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}
