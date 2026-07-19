package models

type CreateSubscriptionRequest struct {
	CustomerID int                  `json:"customer_id"`
	PlanID     int                  `json:"plan_id"`
	AddOns     []CreateSubAddOnItem `json:"add_ons"`
	CouponCode string               `json:"coupon_code,omitempty"`
	Meta       JSONMap              `json:"meta"`
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
	CouponCode string               `json:"coupon_code,omitempty"`
}

type SubscriptionPreviewResponse struct {
	PlanAmount     float64                   `json:"plan_amount"`
	AddOnAmount    float64                   `json:"add_on_amount"`
	DiscountAmount float64                   `json:"discount_amount"`
	TotalAmount    float64                   `json:"total_amount"`
	Items          []SubscriptionPreviewItem `json:"items"`
}

type SubscriptionPreviewItem struct {
	Type      string  `json:"type"`
	Name      string  `json:"name"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
	Subtotal  float64 `json:"subtotal"`
}

type SubscriptionDetailResponse struct {
	ID              int                          `json:"id"`
	CustomerID      int                          `json:"customer_id"`
	PlanID          int                          `json:"plan_id"`
	Status          string                       `json:"status"`
	StartDate       string                       `json:"start_date"`
	NextBillingDate *string                      `json:"next_billing_date"`
	EndDate         *string                      `json:"end_date"`
	TrialEndsAt     *string                      `json:"trial_ends_at"`
	Meta            JSONMap                      `json:"meta"`
	CreatedAt       string                       `json:"created_at"`
	Customer        *SubscriptionCustomerInfo    `json:"customer,omitempty"`
	Plan            *SubscriptionPlanInfo        `json:"plan,omitempty"`
	Product         *SubscriptionProductInfo     `json:"product,omitempty"`
	Service         *SubscriptionServiceInfo     `json:"service,omitempty"`
	AddOns          []SubscriptionAddOnResponse  `json:"add_ons,omitempty"`
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
	ID       int     `json:"id"`
	AddOnID  int     `json:"add_on_id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
}

type SubscriptionCouponResponse struct {
	ID            int     `json:"id"`
	CouponID      int     `json:"coupon_id"`
	Code          string  `json:"code"`
	DiscountType  string  `json:"discount_type"`
	DiscountValue float64 `json:"discount_value"`
}
