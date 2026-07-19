package models

import (
	"time"
)

type SubscriptionCoupon struct {
	ID             int        `db:"id" json:"id"`
	SubscriptionID int        `db:"subscription_id" json:"subscription_id"`
	CouponID       int        `db:"coupon_id" json:"coupon_id"`
	AppliedAt      time.Time  `db:"applied_at" json:"applied_at"`
	ExpiredAt      *time.Time `db:"expired_at" json:"expired_at"`
	CreatedAt      time.Time  `db:"created_at" json:"created_at"`
}
