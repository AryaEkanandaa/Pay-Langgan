package revenue

import (
	"pay-langgan/internal/models"
	"time"
)

func (r *DashboardRepository) BusinessTimezone(businessID string) (string, error) {
	var zone string
	err := r.db.Get(&zone, `SELECT COALESCE(NULLIF(meta->>'timezone',''),'Asia/Makassar') FROM businesses WHERE id=$1`, businessID)
	return zone, err
}

// Uses the existing dashboard status/date definitions, with explicit local-day bounds.
func (r *DashboardRepository) RangeSeries(businessID string, start, end time.Time, zone, granularity string, sales bool) ([]models.DashboardRangePoint, error) {
	const query = `WITH tenant_subs AS (
 SELECT s.id,s.created_at,s.end_date,s.status FROM subscriptions s
 JOIN customers c ON c.id=s.customer_id WHERE c.business_id=$1
 ), events AS (
 SELECT p.paid_at AS occurred_at,p.amount AS revenue,0 AS new_subscriptions,0 AS cancelled_subscriptions,1 AS payments_succeeded,0 AS payments_failed,
 CASE WHEN ac.status='succeeded' THEN 1 ELSE 0 END AS autopay_succeeded,0 AS autopay_failed
 FROM payments p JOIN invoices i ON i.id=p.invoice_id JOIN tenant_subs s ON s.id=i.subscription_id
 LEFT JOIN autopay_charges ac ON ac.payment_id=p.id WHERE p.status='success' AND NOT $6
 UNION ALL
 SELECT p.created_at,0,0,0,0,1,0,0 FROM payments p JOIN invoices i ON i.id=p.invoice_id JOIN tenant_subs s ON s.id=i.subscription_id WHERE p.status='failed' AND NOT $6
 UNION ALL
 SELECT ac.updated_at,0,0,0,0,0,0,1 FROM autopay_charges ac JOIN payments p ON p.id=ac.payment_id JOIN invoices i ON i.id=p.invoice_id JOIN tenant_subs s ON s.id=i.subscription_id WHERE ac.status='failed' AND NOT $6
 UNION ALL
 SELECT created_at,0,1,0,0,0,0,0 FROM tenant_subs
 UNION ALL
 SELECT end_date,0,0,1,0,0,0,0 FROM tenant_subs WHERE status='cancelled'
 ) SELECT TO_CHAR(date_trunc($5,occurred_at AT TIME ZONE 'UTC' AT TIME ZONE $4),'YYYY-MM-DD') AS date,
 SUM(revenue) AS revenue,SUM(new_subscriptions) AS new_subscriptions,SUM(cancelled_subscriptions) AS cancelled_subscriptions,
 SUM(payments_succeeded) AS payments_succeeded,SUM(payments_failed) AS payments_failed,SUM(autopay_succeeded) AS autopay_succeeded,SUM(autopay_failed) AS autopay_failed
 FROM events WHERE occurred_at >= $2 AND occurred_at < $3 GROUP BY 1 ORDER BY 1`
	rows := []models.DashboardRangePoint{}
	err := r.db.Select(&rows, query, businessID, start.UTC(), end.UTC(), zone, granularity, sales)
	return rows, err
}

func (r *DashboardRepository) RangePlanRevenue(businessID string, start, end time.Time) ([]models.PlanRevenuePoint, error) {
	const query = `SELECT pl.id AS plan_id,pl.name,COALESCE(SUM(p.amount),0) AS amount FROM payments p
 JOIN invoices i ON i.id=p.invoice_id JOIN subscriptions s ON s.id=i.subscription_id JOIN customers c ON c.id=s.customer_id JOIN plans pl ON pl.id=s.plan_id
 WHERE c.business_id=$1 AND p.status='success' AND p.paid_at >= $2 AND p.paid_at < $3 GROUP BY pl.id,pl.name ORDER BY amount DESC,pl.id`
	rows := []models.PlanRevenuePoint{}
	err := r.db.Select(&rows, query, businessID, start.UTC(), end.UTC())
	return rows, err
}
