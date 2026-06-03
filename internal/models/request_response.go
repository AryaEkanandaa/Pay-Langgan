package models

import "time"

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Meta    *Pagination `json:"meta,omitempty"`
}

type Pagination struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

type SignupRequest struct {
	BusinessName string `json:"business_name" validate:"required"`
	Name         string `json:"name" validate:"required"`
	Email        string `json:"email" validate:"required,email"`
	Password     string `json:"password" validate:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	Token    string       `json:"token"`
	User     UserDTO      `json:"user"`
	Business BusinessDTO `json:"business"`
}

type LoginResponse struct {
	Token string  `json:"token"`
	User  UserDTO `json:"user"`
}

type MeResponse struct {
	UserID     int    `json:"user_id"`
	BusinessID string `json:"business_id"`
	Email      string `json:"email"`
	Role       string `json:"role"`
}

type UserDTO struct {
	ID         int    `json:"id"`
	BusinessID string `json:"business_id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Role       string `json:"role"`
}

type BusinessDTO struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type UpdateBusinessRequest struct {
	Name string  `json:"name"`
	Meta JSONMap `json:"meta"`
}

type CreateServiceRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Meta        JSONMap `json:"meta"`
}

type UpdateServiceRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Meta        JSONMap `json:"meta"`
}

type CreateProductRequest struct {
	ServiceID   int     `json:"service_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
	Meta        JSONMap `json:"meta"`
}

type UpdateProductRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
	Meta        JSONMap `json:"meta"`
}

type CreatePlanRequest struct {
	ProductID    int     `json:"product_id"`
	Name         string  `json:"name"`
	Price        float64 `json:"price"`
	BillingCycle string  `json:"billing_cycle"`
	TrialDays    int     `json:"trial_days"`
	Meta         JSONMap `json:"meta"`
}

type UpdatePlanRequest struct {
	Name         string   `json:"name"`
	Price        float64  `json:"price"`
	BillingCycle string   `json:"billing_cycle"`
	TrialDays    *int     `json:"trial_days"`
	Meta         JSONMap  `json:"meta"`
}

type CreateAddOnRequest struct {
	ProductID    int     `json:"product_id"`
	Name         string  `json:"name"`
	Price        float64 `json:"price"`
	BillingCycle string  `json:"billing_cycle"`
	Meta         JSONMap `json:"meta"`
}

type UpdateAddOnRequest struct {
	Name         string  `json:"name"`
	Price        float64 `json:"price"`
	BillingCycle string  `json:"billing_cycle"`
	Meta         JSONMap `json:"meta"`
}

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

type CreateCustomerRequest struct {
	Name    string  `json:"name"`
	Email   *string `json:"email"`
	Contact *string `json:"contact"`
	Meta    JSONMap `json:"meta"`
}

type UpdateCustomerRequest struct {
	Name    string  `json:"name"`
	Email   *string `json:"email"`
	Contact *string `json:"contact"`
	Meta    JSONMap `json:"meta"`
}

type CreateSubscriptionRequest struct {
	CustomerID int                  `json:"customer_id"`
	PlanID     int                  `json:"plan_id"`
	AddOns     []CreateSubAddOnItem `json:"add_ons"`
	CouponCode string              `json:"coupon_code,omitempty"`
	Meta       JSONMap             `json:"meta"`
}

type CreateSubAddOnItem struct {
	AddOnID  int `json:"add_on_id"`
	Quantity int `json:"quantity"`
}

type CancelSubscriptionRequest struct {
	Reason string `json:"reason"`
}

type PauseSubscriptionRequest struct {
	Reason string `json:"reason"`
}

type AddSubAddOnRequest struct {
	AddOnID  int `json:"add_on_id"`
	Quantity int `json:"quantity"`
}

type ApplyCouponRequest struct {
	CouponCode string `json:"coupon_code"`
}

type SubscriptionPreviewRequest struct {
	PlanID     int                  `json:"plan_id"`
	AddOns     []CreateSubAddOnItem `json:"add_ons"`
	CouponCode string              `json:"coupon_code,omitempty"`
}

type SubscriptionPreviewResponse struct {
	PlanAmount    float64                    `json:"plan_amount"`
	AddOnAmount   float64                    `json:"add_on_amount"`
	DiscountAmount float64                   `json:"discount_amount"`
	TotalAmount   float64                    `json:"total_amount"`
	Items         []SubscriptionPreviewItem  `json:"items"`
}

type SubscriptionPreviewItem struct {
	Type      string  `json:"type"`
	Name      string  `json:"name"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
	Subtotal  float64 `json:"subtotal"`
}

type SubscriptionDetailResponse struct {
	ID              int                         `json:"id"`
	CustomerID      int                         `json:"customer_id"`
	PlanID          int                         `json:"plan_id"`
	Status          string                      `json:"status"`
	StartDate       string                      `json:"start_date"`
	NextBillingDate *string                     `json:"next_billing_date"`
	EndDate         *string                     `json:"end_date"`
	TrialEndsAt     *string                     `json:"trial_ends_at"`
	Meta            JSONMap                     `json:"meta"`
	CreatedAt       string                      `json:"created_at"`
	Customer        *SubscriptionCustomerInfo   `json:"customer,omitempty"`
	Plan            *SubscriptionPlanInfo       `json:"plan,omitempty"`
	Product         *SubscriptionProductInfo    `json:"product,omitempty"`
	Service         *SubscriptionServiceInfo    `json:"service,omitempty"`
	AddOns          []SubscriptionAddOnResponse `json:"add_ons,omitempty"`
	Coupons         []SubscriptionCouponResponse `json:"coupons,omitempty"`
}

type SubscriptionCustomerInfo struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Email *string `json:"email"`
}

type SubscriptionPlanInfo struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	Price        float64 `json:"price"`
	BillingCycle string  `json:"billing_cycle"`
	TrialDays    int     `json:"trial_days"`
}

type SubscriptionProductInfo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type SubscriptionServiceInfo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type SubscriptionAddOnResponse struct {
	ID        int     `json:"id"`
	AddOnID   int     `json:"add_on_id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
}

type SubscriptionCouponResponse struct {
	ID            int     `json:"id"`
	CouponID      int     `json:"coupon_id"`
	Code          string  `json:"code"`
	DiscountType  string  `json:"discount_type"`
	DiscountValue float64 `json:"discount_value"`
}


