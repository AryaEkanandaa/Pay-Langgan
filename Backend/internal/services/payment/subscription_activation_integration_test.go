package payment_test

import (
	"context"
	billing "pay-langgan/internal/services/billing"
	"testing"
)

func TestSubscriptionActivationPostgres(t *testing.T) {
	db := paymentTestDB(t)
	ctx := context.Background()
	for _, status := range []string{"pending", "paused", "cancelled", "trial", "active"} {
		tx, err := db.BeginTxx(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = tx.Exec(`UPDATE subscriptions SET status=$1 WHERE id=1`, status); err != nil {
			t.Fatal(err)
		}
		if err = billing.ActivatePaidSubscription(ctx, tx, 1); err != nil {
			t.Fatal(err)
		}
		var got string
		if err = tx.Get(&got, `SELECT status FROM subscriptions WHERE id=1`); err != nil {
			t.Fatal(err)
		}
		if got != status {
			t.Fatalf("unpaid invoice activated %s", status)
		}
		if _, err = tx.Exec(`UPDATE invoices SET status='paid' WHERE id=1`); err != nil {
			t.Fatal(err)
		}
		if err = billing.ActivatePaidSubscription(ctx, tx, 1); err != nil {
			t.Fatal(err)
		}
		if err = tx.Get(&got, `SELECT status FROM subscriptions WHERE id=1`); err != nil {
			t.Fatal(err)
		}
		want := status
		if status == "pending" {
			want = "active"
		}
		if got != want {
			t.Fatalf("payment state %s got %s", status, got)
		}
		if err = tx.Rollback(); err != nil {
			t.Fatal(err)
		}
	}
}
