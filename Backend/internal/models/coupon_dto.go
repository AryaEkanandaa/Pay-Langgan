package models

import "time"

type CreateCouponRequest struct {
	Code          string     `json:"code"`
	DiscountType  string     `json:"discount_type"`
	DiscountValue float64    `json:"discount_value"`
	MaxUsage      *int       `json:"max_usage"`
	ExpiresAt     *time.Time `json:"expires_at"`
}

type UpdateCouponRequest struct {
	Code          string     `json:"code"`
	DiscountType  string     `json:"discount_type"`
	DiscountValue float64    `json:"discount_value"`
	MaxUsage      *int       `json:"max_usage"`
	ExpiresAt     *time.Time `json:"expires_at"`
}
