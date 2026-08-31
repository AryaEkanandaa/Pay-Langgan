package models

type DashboardSummary struct {
	ActiveCustomers     int                       `json:"active_customers"`
	ActiveSubscriptions int                       `json:"active_subscriptions"`
	MRR                 float64                   `json:"mrr"`
	DueInvoices         int                       `json:"due_invoices"`
	MonthlyPerformance  []MonthlyPerformancePoint `json:"monthly_performance"`
	YearlyBreakdown     []YearlyBreakdownPoint    `json:"yearly_breakdown"`
	IncomeSharePct      int                       `json:"income_share_pct"`
	SpendingSharePct    int                       `json:"spending_share_pct"`
}

type MonthlyPerformancePoint struct {
	Month string  `json:"month"`
	Value float64 `json:"value"`
}

type YearlyBreakdownPoint struct {
	Year     string  `json:"year"`
	Income   float64 `json:"income"`
	Spending float64 `json:"spending"`
}
