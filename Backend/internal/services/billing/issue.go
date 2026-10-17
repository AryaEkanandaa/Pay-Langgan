package billing

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"pay-langgan/internal/models"
	billingrepo "pay-langgan/internal/repositories/billing"
	"pay-langgan/internal/utils"
)

type charge struct {
	Kind     string `db:"kind"`
	Name     string `db:"name"`
	Quantity int64  `db:"quantity"`
	Cents    int64  `db:"cents"`
	Cycle    string `db:"cycle"`
}

// QuoteCharge and PriceQuote share invoice rounding and discount rules with checkout previews.
type QuoteCharge = charge

func PriceQuote(charges []QuoteCharge, cycle string) ([]int64, int64, error) {
	return priceCharges(charges, cycle)
}

const maxInvoiceCents int64 = 999999999999 // NUMERIC(12,2)

// priceCharges works in hundredths, avoiding floating-point discount rounding.
// Discounts apply sequentially to the remaining balance in coupon ID order.
func priceCharges(charges []charge, cycle string) ([]int64, int64, error) {
	subtotals := make([]int64, len(charges))
	var total int64
	for i, c := range charges {
		if c.Cents < 0 || c.Cents > maxInvoiceCents || c.Quantity < 1 {
			return nil, 0, fmt.Errorf("invalid charge %q", c.Name)
		}
		switch c.Kind {
		case "plan", "add_on":
			if c.Cycle != cycle {
				return nil, 0, fmt.Errorf("billing cycle mismatch for %q", c.Name)
			}
			if c.Cents > 0 && c.Quantity > (maxInvoiceCents-total)/c.Cents {
				return nil, 0, fmt.Errorf("invoice amount exceeds storage limit")
			}
			subtotals[i] = c.Cents * c.Quantity
		case "percentage", "fixed":
			discount := c.Cents
			if c.Kind == "percentage" {
				if c.Cents > 10000 {
					return nil, 0, fmt.Errorf("percentage exceeds 100")
				}
				discount = total * c.Cents / 10000
			}
			if discount > total {
				discount = total
			}
			subtotals[i] = -discount
		default:
			return nil, 0, fmt.Errorf("unknown charge type %q", c.Kind)
		}
		total += subtotals[i]
	}
	return subtotals, total, nil
}

// IssueInvoice must run in the transaction owning the subscription row lock
// (or the transaction creating that subscription). The snapshot uses one query.
func IssueInvoice(tx *sqlx.Tx, repo *billingrepo.InvoiceRepository, subscriptionID int, start, end time.Time, cycle string, initial bool) (*models.Invoice, error) {
	return IssueInvoiceContext(context.Background(), tx, repo, subscriptionID, start, end, cycle, initial)
}

func IssueInvoiceContext(ctx context.Context, tx *sqlx.Tx, repo *billingrepo.InvoiceRepository, subscriptionID int, start, end time.Time, cycle string, initial bool) (*models.Invoice, error) {
	if !end.After(start) {
		return nil, fmt.Errorf("invalid invoice period")
	}
	var charges []charge
	err := tx.SelectContext(ctx, &charges, `
 WITH owner AS (
  SELECT s.id, s.plan_id, c.business_id FROM subscriptions s
  JOIN customers c ON c.id = s.customer_id WHERE s.id = $1
 ), charges AS (
  SELECT 0 AS position, p.id AS ordering, 'plan' AS kind, p.name, 1::bigint AS quantity,
   (p.price*100)::bigint AS cents, p.billing_cycle AS cycle
  FROM owner o JOIN plans p ON p.id = o.plan_id
  UNION ALL
  SELECT 1, a.id, 'add_on', a.name, sa.quantity::bigint, (a.price*100)::bigint, a.billing_cycle
  FROM owner o JOIN subscription_add_ons sa ON sa.subscription_id = o.id JOIN add_ons a ON a.id = sa.add_on_id
  UNION ALL
  SELECT 2, c.id, c.discount_type, 'Coupon ' || c.code, 1::bigint, (c.discount_value*100)::bigint, ''
  FROM owner o JOIN subscription_coupons sc ON sc.subscription_id = o.id
  JOIN coupons c ON c.id = sc.coupon_id AND c.business_id = o.business_id
  WHERE ($3 OR sc.applied_at <= $2)
   AND (sc.expired_at IS NULL OR sc.expired_at > $2)
   AND (c.expires_at IS NULL OR c.expires_at > $2)
 ) SELECT kind, name, quantity, cents, cycle FROM charges ORDER BY position, ordering`, subscriptionID, start, initial)
	if err != nil {
		return nil, fmt.Errorf("load invoice snapshot: %w", err)
	}
	if len(charges) == 0 || charges[0].Kind != "plan" {
		return nil, fmt.Errorf("missing subscription plan")
	}
	// The subscription fixes the base plan's cadence at registration. Add-ons
	// must match it; changing a catalog cycle cannot silently reschedule billing.
	charges[0].Cycle = cycle
	subtotals, total, err := priceCharges(charges, cycle)
	if err != nil {
		return nil, err
	}
	invoice := &models.Invoice{SubscriptionID: subscriptionID, InvoiceNumber: utils.GenerateInvoiceNumber(), Amount: float64(total) / 100, Status: "pending", DueDate: &start, PeriodStart: &start, PeriodEnd: &end}
	if total == 0 {
		invoice.Status = "paid"
	}
	if err := repo.CreateContext(ctx, tx, invoice); err != nil {
		return nil, err
	}
	for i, c := range charges {
		kind := c.Kind
		unit := c.Cents
		if kind == "percentage" || kind == "fixed" {
			kind = "discount"
			unit = subtotals[i]
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO invoice_items (invoice_id, item_type, description, quantity, unit_price, subtotal)
    VALUES ($1,$2,$3,$4,$5::numeric/100,$6::numeric/100)`, invoice.ID, kind, c.Name, c.Quantity, unit, subtotals[i])
		if err != nil {
			return nil, fmt.Errorf("save invoice snapshot: %w", err)
		}
	}
	if total == 0 {
		if err := tx.QueryRowContext(ctx, `UPDATE invoices SET paid_at = NOW() WHERE id = $1 RETURNING paid_at`, invoice.ID).Scan(&invoice.PaidAt); err != nil {
			return nil, err
		}
	}
	if total > 0 {
		if err := queueInvoiceEmail(ctx, tx, invoice.ID, false); err != nil {
			return nil, fmt.Errorf("queue invoice email: %w", err)
		}
	}
	if total == 0 {
		if err := ActivatePaidSubscription(ctx, tx, invoice.ID); err != nil {
			return nil, err
		}
		if err := QueuePaymentConfirmation(ctx, tx, invoice.ID); err != nil {
			return nil, err
		}
	}
	return invoice, nil
}
