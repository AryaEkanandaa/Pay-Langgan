package billing

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"pay-langgan/internal/database"
	billingrepo "pay-langgan/internal/repositories/billing"
)

type RenewalService struct {
	db       *database.DB
	invoices *billingrepo.InvoiceRepository
}

func NewRenewalService(db *database.DB) *RenewalService {
	return &RenewalService{db: db, invoices: billingrepo.NewInvoiceRepository(db)}
}

type RenewalResult struct {
	InvoicesCreated int              `json:"invoices_created"`
	LegacySkipped   int              `json:"legacy_skipped"`
	Failures        []RenewalFailure `json:"failures"`
}
type RenewalFailure struct {
	SubscriptionID int    `json:"subscription_id"`
	Error          string `json:"error"`
}

// Run issues at most limit periods, including missed periods, in independent
// transactions. Failed subscriptions are skipped for the remainder of this run.
func (s *RenewalService) Run(ctx context.Context, at time.Time, limit int) (RenewalResult, error) {
	result := RenewalResult{Failures: []RenewalFailure{}}
	if at.IsZero() || limit < 1 || limit > 10000 {
		return result, fmt.Errorf("time is required and limit must be 1..10000")
	}
	at = at.UTC()
	err := s.db.GetContext(ctx, &result.LegacySkipped, `SELECT count(*) FROM subscriptions
  WHERE status IN ('active','trial') AND next_billing_date <= $1 AND billing_anchor IS NULL`, at)
	if err != nil {
		return result, err
	}
	skipped := []int{}
	for attempt := 0; attempt < limit; attempt++ {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		id, found, err := s.renewOne(ctx, at, skipped)
		if err != nil {
			if id == 0 {
				return result, err
			}
			skipped = append(skipped, id)
			result.Failures = append(result.Failures, RenewalFailure{id, err.Error()})
			continue
		}
		if !found {
			break
		}
		result.InvoicesCreated++
	}
	return result, nil
}

func (s *RenewalService) renewOne(ctx context.Context, at time.Time, skipped []int) (int, bool, error) {
	tx, err := s.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return 0, false, err
	}
	defer tx.Rollback()
	query := `SELECT id, status, next_billing_date, billing_anchor, billing_cycle FROM subscriptions
  WHERE status IN ('active','trial') AND next_billing_date <= ? AND billing_anchor IS NOT NULL`
	args := []interface{}{at}
	if len(skipped) > 0 {
		query += ` AND id NOT IN (?)`
		args = append(args, skipped)
	}
	query += ` ORDER BY next_billing_date, id LIMIT 1 FOR UPDATE SKIP LOCKED`
	query, args, err = sqlx.In(query, args...)
	if err != nil {
		return 0, false, err
	}
	var sub struct {
		Status string    `db:"status"`
		ID     int       `db:"id"`
		Start  time.Time `db:"next_billing_date"`
		Anchor time.Time `db:"billing_anchor"`
		Cycle  string    `db:"billing_cycle"`
	}
	err = tx.GetContext(ctx, &sub, tx.Rebind(query), args...)
	if err == sql.ErrNoRows {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	end, err := NextPeriod(sub.Anchor, sub.Start, sub.Cycle)
	if err != nil {
		return sub.ID, false, err
	}
	invoice, err := IssueInvoiceContext(ctx, tx, s.invoices, sub.ID, sub.Start, end, sub.Cycle, false)
	if err != nil {
		return sub.ID, false, err
	}
	// Trial orders have no invoice at registration. Attach their first bill for order status/retries.
	if _, err = tx.ExecContext(ctx, `UPDATE merchant_orders SET invoice_id=$2 WHERE subscription_id=$1 AND invoice_id IS NULL`, sub.ID, invoice.ID); err != nil {
		return sub.ID, false, err
	}
	status := "active"
	if sub.Status == "trial" && invoice.Status != "paid" {
		status = "pending"
	}
	_, err = tx.ExecContext(ctx, `UPDATE subscriptions SET status = $3, next_billing_date = $2 WHERE id = $1`, sub.ID, end, status)
	if err != nil {
		return sub.ID, false, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO audit_logs (business_id, action, entity_type, entity_id, new_value)
  SELECT c.business_id, 'renew_subscription', 'subscription', s.id::text,
   jsonb_build_object('invoice_id', $2::int, 'period_start', $3::text, 'period_end', $4::text)
  FROM subscriptions s JOIN customers c ON c.id = s.customer_id WHERE s.id = $1`, sub.ID, invoice.ID, sub.Start.Format(time.RFC3339Nano), end.Format(time.RFC3339Nano))
	if err != nil {
		return sub.ID, false, err
	}
	if err := tx.Commit(); err != nil {
		return sub.ID, false, err
	}
	return sub.ID, true, nil
}
