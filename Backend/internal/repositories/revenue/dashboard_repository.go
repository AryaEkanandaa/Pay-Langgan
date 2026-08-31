package revenue

import (
	"fmt"

	"pay-langgan/internal/database"
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
		  AND s.status IN ('trial', 'active')`
	var mrr float64
	if err := r.db.Get(&mrr, query, businessID); err != nil {
		return 0, fmt.Errorf("calculate mrr: %w", err)
	}
	return mrr, nil
}

func (r *DashboardRepository) CountDueInvoices(businessID string) (int, error) {
	query := `SELECT COUNT(*)
		FROM invoices i
		JOIN subscriptions s ON s.id = i.subscription_id
		JOIN customers c ON c.id = s.customer_id
		WHERE c.business_id = $1
		  AND i.status = 'pending'
		  AND i.due_date IS NOT NULL
		  AND i.due_date <= NOW()`
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
		  AND p.paid_at >= DATE_TRUNC('month', NOW()) - INTERVAL '6 months'
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
		  AND p.paid_at >= DATE_TRUNC('year', NOW()) - INTERVAL '4 years'
		GROUP BY EXTRACT(YEAR FROM p.paid_at)
		ORDER BY EXTRACT(YEAR FROM p.paid_at)`
	var rows []YearlyRevenuePoint
	if err := r.db.Select(&rows, query, businessID); err != nil {
		return nil, fmt.Errorf("load yearly revenue: %w", err)
	}
	return rows, nil
}
