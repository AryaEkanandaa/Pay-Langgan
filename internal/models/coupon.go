package models

import (
	"time"
)

type Coupon struct {
	ID            int          `db:"id" json:"id"`
	Code          string       `db:"code" json:"code"`
	DiscountType  string       `db:"discount_type" json:"discount_type"`
	DiscountValue float64      `db:"discount_value" json:"discount_value"`
	MaxUsage      *int         `db:"max_usage" json:"max_usage"`
	UsedCount     int          `db:"used_count" json:"used_count"`
	ExpiresAt     *time.Time   `db:"expires_at" json:"expires_at"`
	CreatedAt     time.Time    `db:"created_at" json:"created_at"`
}

func (c *Coupon) IsValid() bool {
	if c.MaxUsage != nil && c.UsedCount >= *c.MaxUsage {
		return false
	}
	if c.ExpiresAt != nil && c.ExpiresAt.Before(time.Now()) {
		return false
	}
	return true
}
