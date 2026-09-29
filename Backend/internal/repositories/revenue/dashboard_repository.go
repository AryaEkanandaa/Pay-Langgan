package revenue

import (
	"fmt"

	"pay-langgan/internal/database"
	"pay-langgan/internal/models"
)

type DashboardRepository struct {
	db *database.DB
}

type MonthlyRevenuePoint struct {
	Month string  `db:"month"`
	Value float64 `db:"value"`
}

type YearlyRevenuePoint struct {
	Year   string  `db:"year"`
	Income float64 `db:"income"`
}

func NewDashboardRepository(db *database.DB) *DashboardRepository {
	return &DashboardRepository{db: db}
}

func (r *DashboardRepository) CountActiveCustomers(businessID string) (int, error) {
	query := `SELECT COUNT(DISTINCT c.id)
		FROM customers c
		JOIN subscriptions s ON s.customer_id = c.id
		WHERE c.business_id = $1 AND s.status IN ('trial', 'active')`
	var count int
	if err := r.db.Get(&count, query, businessID); err != nil {
		return 0, fmt.Errorf("count active customers: %w", err)
	}
	return count, nil
}

func (r *DashboardRepository) CountActiveSubscriptions(businessID string) (int, error) {
	query := `SELECT COUNT(*)
		FROM subscriptions s
		JOIN customers c ON c.id = s.customer_id
		WHERE c.business_id = $1 AND s.status IN ('trial', 'active')`
	var count int
	if err := r.db.Get(&count, query, businessID); err != nil {
		return 0, fmt.Errorf("count active subscriptions: %w", err)
	}
	return count, nil
}

func (r *DashboardRepository) CalculateMRR(businessID string) (float64, error) {
	query := `SELECT COALESCE(SUM(
			CASE WHEN pl.billing_cycle = 'yearly' THEN pl.price / 12 ELSE pl.price END
			+ COALESCE(addons.amount, 0)
		), 0)
		FROM subscriptions s
		JOIN customers c ON c.id = s.customer_id
		JOIN plans pl ON pl.id = s.plan_id
		JOIN products pr ON pr.id = pl.product_id
		JOIN services sv ON sv.id = pr.service_id
		LEFT JOIN (
			SELECT sao.subscription_id,
				SUM(CASE WHEN ao.billing_cycle = 'yearly'
					THEN ao.price * sao.quantity / 12
					ELSE ao.price * sao.quantity END) AS amount
			FROM subscription_add_ons sao
			JOIN add_ons ao ON ao.id = sao.add_on_id
			GROUP BY sao.subscription_id
		) addons ON addons.subscription_id = s.id
		WHERE c.business_id = $1
		  AND sv.business_id = $1
		  AND s.status = 'active'`
	var mrr float64
	if err := r.db.Get(&mrr, query, businessID); err != nil {
		return 0, fmt.Errorf("calculate mrr: %w", err)
	}
	return mrr, nil
}

type TenantFinancials struct {
	RevenueThisMonth       float64 `db:"revenue_this_month"`
	OutstandingAmount      float64 `db:"outstanding_amount"`
	OverdueAmount          float64 `db:"overdue_amount"`
	PaymentsSucceeded      int     `db:"payments_succeeded"`
	PaymentsFailed         int     `db:"payments_failed"`
	AutopaySucceeded       int     `db:"autopay_succeeded"`
	AutopayFailed          int     `db:"autopay_failed"`
	NewSubscriptions       int     `db:"new_subscriptions"`
	CancelledSubscriptions int     `db:"cancelled_subscriptions"`
}

// Financials uses UTC calendar months, matching the payment and invoice timestamps.
func (r *DashboardRepository) Financials(businessID string) (TenantFinancials, error) {
	var result TenantFinancials
	const query = `WITH bounds AS (
 SELECT date_trunc('month', now() AT TIME ZONE 'UTC') AS month_start,
        date_trunc('month', now() AT TIME ZONE 'UTC') + interval '1 month' AS next_month,
        now() AT TIME ZONE 'UTC' AS current_time
), tenant_subs AS (
 SELECT s.id, s.created_at, s.end_date, s.status FROM subscriptions s
 JOIN customers c ON c.id=s.customer_id WHERE c.business_id=$1
), tenant_invoices AS (
 SELECT i.id,i.amount,i.status,i.due_date FROM invoices i
 JOIN tenant_subs s ON s.id=i.subscription_id
), payment_stats AS (
 SELECT COALESCE(SUM(p.amount) FILTER (WHERE p.status='success' AND p.paid_at >= b.month_start AND p.paid_at < b.next_month),0) AS revenue_this_month,
 COUNT(*) FILTER (WHERE p.status='success' AND p.paid_at >= b.month_start AND p.paid_at < b.next_month) AS payments_succeeded,
 COUNT(*) FILTER (WHERE p.status='failed' AND p.created_at >= b.month_start AND p.created_at < b.next_month) AS payments_failed,
 COUNT(*) FILTER (WHERE ac.status='succeeded' AND p.paid_at >= b.month_start AND p.paid_at < b.next_month) AS autopay_succeeded,
 COUNT(*) FILTER (WHERE ac.status='failed' AND ac.updated_at >= b.month_start AND ac.updated_at < b.next_month) AS autopay_failed
 FROM payments p JOIN tenant_invoices i ON i.id=p.invoice_id
 LEFT JOIN autopay_charges ac ON ac.payment_id=p.id CROSS JOIN bounds b
), invoice_stats AS (
 SELECT COALESCE(SUM(i.amount) FILTER (WHERE i.status='pending'),0) AS outstanding_amount,
 COALESCE(SUM(i.amount) FILTER (WHERE i.status='pending' AND i.due_date <= b.current_time),0) AS overdue_amount
 FROM tenant_invoices i CROSS JOIN bounds b
), subscription_stats AS (
 SELECT COUNT(*) FILTER (WHERE s.created_at >= b.month_start AND s.created_at < b.next_month) AS new_subscriptions,
 COUNT(*) FILTER (WHERE s.status='cancelled' AND s.end_date >= b.month_start AND s.end_date < b.next_month) AS cancelled_subscriptions
 FROM tenant_subs s CROSS JOIN bounds b
)
 SELECT * FROM payment_stats CROSS JOIN invoice_stats CROSS JOIN subscription_stats`
	if err := r.db.Get(&result, query, businessID); err != nil {
		return result, fmt.Errorf("load tenant financials: %w", err)
	}
	return result, nil
}

func (r *DashboardRepository) RevenueByPlan(businessID string) ([]models.PlanRevenuePoint, error) {
	const query = `SELECT pl.id AS plan_id, pl.name,
 COALESCE(SUM(p.amount),0) AS amount FROM payments p
 JOIN invoices i ON i.id=p.invoice_id
 JOIN subscriptions s ON s.id=i.subscription_id
 JOIN customers c ON c.id=s.customer_id
 JOIN plans pl ON pl.id=s.plan_id
 WHERE c.business_id=$1 AND p.status='success'
 AND p.paid_at >= date_trunc('month',now() AT TIME ZONE 'UTC')
 AND p.paid_at < date_trunc('month',now() AT TIME ZONE 'UTC')+interval '1 month'
 GROUP BY pl.id,pl.name ORDER BY amount DESC,pl.id LIMIT 5`
	rows := []models.PlanRevenuePoint{}
	if err := r.db.Select(&rows, query, businessID); err != nil {
		return nil, fmt.Errorf("load plan revenue: %w", err)
	}
	return rows, nil
}

func (r *DashboardRepository) PlatformSummary() (models.PlatformDashboardSummary, error) {
	const query = `SELECT
 (SELECT COUNT(*) FROM platform_subscriptions s JOIN businesses b ON b.id=s.business_id WHERE s.status='active' AND NOT b.deleted) AS active_businesses,
 (SELECT COUNT(*) FROM platform_subscriptions s JOIN businesses b ON b.id=s.business_id WHERE s.status='past_due' AND NOT b.deleted) AS past_due_businesses,
 (SELECT COALESCE(SUM(amount),0) FROM platform_invoices WHERE status='paid' AND paid_at >= date_trunc('month',now() AT TIME ZONE 'UTC') AND paid_at < date_trunc('month',now() AT TIME ZONE 'UTC')+interval '1 month') AS revenue_this_month,
 (SELECT COALESCE(SUM(amount),0) FROM platform_invoices WHERE status='pending') AS outstanding_amount,
 (SELECT COUNT(*) FROM platform_autopay_charges WHERE status='failed' AND updated_at >= date_trunc('month',now() AT TIME ZONE 'UTC') AND updated_at < date_trunc('month',now() AT TIME ZONE 'UTC')+interval '1 month') AS failed_autopay_this_month`
	var result models.PlatformDashboardSummary
	if err := r.db.Get(&result, query); err != nil {
		return result, fmt.Errorf("load platform dashboard: %w", err)
	}
	return result, nil
}

func (r *DashboardRepository) CountDueInvoices(businessID string) (int, error) {
	query := `SELECT COUNT(*)
		FROM invoices i
		JOIN subscriptions s ON s.id = i.subscription_id
		JOIN customers c ON c.id = s.customer_id
		WHERE c.business_id = $1
		  AND i.status = 'pending'
		  AND i.due_date IS NOT NULL
		  AND i.due_date <= (NOW() AT TIME ZONE 'UTC')`
	var count int
	if err := r.db.Get(&count, query, businessID); err != nil {
		return 0, fmt.Errorf("count due invoices: %w", err)
	}
	return count, nil
}

func (r *DashboardRepository) MonthlyRevenue(businessID string) ([]MonthlyRevenuePoint, error) {
	query := `SELECT TO_CHAR(DATE_TRUNC('month', p.paid_at), 'YYYY-MM') AS month,
			COALESCE(SUM(p.amount), 0) AS value
		FROM payments p
		JOIN invoices i ON i.id = p.invoice_id
		JOIN subscriptions s ON s.id = i.subscription_id
		JOIN customers c ON c.id = s.customer_id
		WHERE c.business_id = $1
		  AND p.status = 'success'
		  AND p.paid_at >= DATE_TRUNC('month', NOW() AT TIME ZONE 'UTC') - INTERVAL '6 months'
		GROUP BY DATE_TRUNC('month', p.paid_at)
		ORDER BY DATE_TRUNC('month', p.paid_at)`
	var rows []MonthlyRevenuePoint
	if err := r.db.Select(&rows, query, businessID); err != nil {
		return nil, fmt.Errorf("load monthly revenue: %w", err)
	}
	return rows, nil
}

func (r *DashboardRepository) YearlyRevenue(businessID string) ([]YearlyRevenuePoint, error) {
	query := `SELECT EXTRACT(YEAR FROM p.paid_at)::TEXT AS year,
			COALESCE(SUM(p.amount), 0) AS income
		FROM payments p
		JOIN invoices i ON i.id = p.invoice_id
		JOIN subscriptions s ON s.id = i.subscription_id
		JOIN customers c ON c.id = s.customer_id
		WHERE c.business_id = $1
		  AND p.status = 'success'
		  AND p.paid_at >= DATE_TRUNC('year', NOW() AT TIME ZONE 'UTC') - INTERVAL '4 years'
		GROUP BY EXTRACT(YEAR FROM p.paid_at)
		ORDER BY EXTRACT(YEAR FROM p.paid_at)`
	var rows []YearlyRevenuePoint
	if err := r.db.Select(&rows, query, businessID); err != nil {
		return nil, fmt.Errorf("load yearly revenue: %w", err)
	}
	return rows, nil
}
