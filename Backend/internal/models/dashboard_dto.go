package models

type DashboardSummary struct {
	RangeStats             *DashboardRangeStats      `json:"range_stats,omitempty"`
	ActiveCustomers        int                       `json:"active_customers"`
	ActiveSubscriptions    int                       `json:"active_subscriptions"`
	MRR                    float64                   `json:"mrr"`
	DueInvoices            int                       `json:"due_invoices"`
	RevenueThisMonth       float64                   `json:"revenue_this_month"`
	OutstandingAmount      float64                   `json:"outstanding_amount"`
	OverdueAmount          float64                   `json:"overdue_amount"`
	PaymentsSucceeded      int                       `json:"payments_succeeded"`
	PaymentsFailed         int                       `json:"payments_failed"`
	AutopaySucceeded       int                       `json:"autopay_succeeded"`
	AutopayFailed          int                       `json:"autopay_failed"`
	NewSubscriptions       int                       `json:"new_subscriptions"`
	CancelledSubscriptions int                       `json:"cancelled_subscriptions"`
	RevenueByPlan          []PlanRevenuePoint        `json:"revenue_by_plan"`
	MonthlyPerformance     []MonthlyPerformancePoint `json:"monthly_performance"`
	YearlyBreakdown        []YearlyBreakdownPoint    `json:"yearly_breakdown"`
}

type PlanRevenuePoint struct {
	PlanID int     `db:"plan_id" json:"plan_id"`
	Name   string  `db:"name" json:"name"`
	Amount float64 `db:"amount" json:"amount"`
}

type PlatformDashboardSummary struct {
	ActiveBusinesses       int     `db:"active_businesses" json:"active_businesses"`
	PastDueBusinesses      int     `db:"past_due_businesses" json:"past_due_businesses"`
	RevenueThisMonth       float64 `db:"revenue_this_month" json:"revenue_this_month"`
	OutstandingAmount      float64 `db:"outstanding_amount" json:"outstanding_amount"`
	FailedAutopayThisMonth int     `db:"failed_autopay_this_month" json:"failed_autopay_this_month"`
}

type MonthlyPerformancePoint struct {
	Month string  `json:"month"`
	Value float64 `json:"value"`
}

type YearlyBreakdownPoint struct {
	Year   string  `json:"year"`
	Income float64 `json:"income"`
}
