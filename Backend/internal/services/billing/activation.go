package billing

import (
	"context"
	"github.com/jmoiron/sqlx"
)

// ActivatePaidSubscription runs in the verified payment transaction. Conditional
// update serializes with cancel/pause and cannot revive those terminal decisions.
// Invoice periods remain the agreed snapshot; payment does not rewrite prices/dates.
func ActivatePaidSubscription(ctx context.Context, tx *sqlx.Tx, invoiceID int) error {
	_, err := tx.ExecContext(ctx, `UPDATE subscriptions s SET status='active'
 FROM invoices i WHERE i.id=$1 AND i.subscription_id=s.id AND i.status='paid' AND s.status='pending'`, invoiceID)
	return err
}
