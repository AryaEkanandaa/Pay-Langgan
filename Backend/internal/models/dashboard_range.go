package models

type DashboardRangePoint struct {
	Date                   string  `json:"date" db:"date"`
	Revenue                float64 `json:"revenue" db:"revenue"`
	NewSubscriptions       int     `json:"new_subscriptions" db:"new_subscriptions"`
	CancelledSubscriptions int     `json:"cancelled_subscriptions" db:"cancelled_subscriptions"`
	PaymentsSucceeded      int     `json:"payments_succeeded" db:"payments_succeeded"`
	PaymentsFailed         int     `json:"payments_failed" db:"payments_failed"`
	AutopaySucceeded       int     `json:"autopay_succeeded" db:"autopay_succeeded"`
	AutopayFailed          int     `json:"autopay_failed" db:"autopay_failed"`
}

type DashboardRangeStats struct {
	From          string                `json:"from"`
	To            string                `json:"to"`
	Timezone      string                `json:"timezone"`
	Granularity   string                `json:"granularity"`
	Series        []DashboardRangePoint `json:"series"`
	RevenueByPlan []PlanRevenuePoint    `json:"revenue_by_plan"`
}
