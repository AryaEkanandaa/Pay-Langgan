package payment

import (
	"context"
	"fmt"
	"time"

	"pay-langgan/internal/utils"
)

type HistoryItem struct {
	ID              string     `db:"id" json:"id"`
	InvoiceID       int        `db:"invoice_id" json:"invoice_id"`
	InvoiceNumber   string     `db:"invoice_number" json:"invoice_number"`
	SubscriptionID  int        `db:"subscription_id" json:"subscription_id"`
	CustomerID      int        `db:"customer_id" json:"customer_id"`
	PaymentID       *int       `db:"payment_id" json:"payment_id"`
	CheckoutID      *int64     `db:"checkout_id" json:"checkout_id"`
	AutopayChargeID *int64     `db:"autopay_charge_id" json:"autopay_charge_id"`
	Provider        string     `db:"provider" json:"provider"`
	Amount          float64    `db:"amount" json:"amount"`
	Status          string     `db:"status" json:"status"`
	SessionStatus   *string    `db:"session_status" json:"session_status"`
	FailureCode     *string    `db:"failure_code" json:"failure_code"`
	ReviewRequired  bool       `db:"review_required" json:"review_required"`
	CreatedAt       time.Time  `db:"created_at" json:"attempted_at"`
	PaidAt          *time.Time `db:"paid_at" json:"paid_at"`
}

const historyRows = `WITH attempts AS (
 SELECT 'checkout:'||cs.id::text AS id,cs.invoice_id,cs.payment_id,cs.id AS checkout_id,NULL::bigint AS autopay_charge_id,cs.provider,
 COALESCE(p.amount,i.amount) AS amount,COALESCE(p.status,cs.status) AS status,cs.status AS session_status,
 cs.last_error AS failure_code,cs.review_required,cs.created_at,p.paid_at
 FROM checkout_sessions cs JOIN invoices i ON i.id=cs.invoice_id LEFT JOIN payments p ON p.id=cs.payment_id
 UNION ALL
 SELECT 'payment:'||p.id::text,p.invoice_id,p.id,NULL,ac.id,p.provider,p.amount,p.status,ac.status,
 ac.failure_code,COALESCE(ac.status='review_required',FALSE),p.created_at,p.paid_at
 FROM payments p LEFT JOIN autopay_charges ac ON ac.payment_id=p.id
 WHERE NOT EXISTS(SELECT 1 FROM checkout_sessions cs WHERE cs.payment_id=p.id)
), history AS (
 SELECT a.*,i.invoice_number,i.subscription_id,c.id AS customer_id FROM attempts a
 JOIN invoices i ON i.id=a.invoice_id JOIN subscriptions s ON s.id=i.subscription_id
 JOIN customers c ON c.id=s.customer_id WHERE c.business_id=$1
 AND ($2::int=0 OR i.id=$2) AND ($3::int=0 OR s.id=$3) AND ($4::text='' OR a.status=$4)
) `

func (s *CheckoutService) PaymentHistory(ctx context.Context, businessID string, invoiceID, subscriptionID, page, limit int, status string) ([]HistoryItem, int, error) {
	if businessID == "" || invoiceID < 0 || subscriptionID < 0 || page < 1 || limit < 1 || limit > 100 {
		return nil, 0, utils.ErrBadRequest
	}
	switch status {
	case "", "pending", "success", "failed", "expired", "created", "creating", "active", "completed":
	default:
		return nil, 0, utils.ErrBadRequest
	}
	args := []any{businessID, invoiceID, subscriptionID, status}
	var total int
	if err := s.db.GetContext(ctx, &total, historyRows+`SELECT count(*) FROM history`, args...); err != nil {
		return nil, 0, err
	}
	items := []HistoryItem{}
	if err := s.db.SelectContext(ctx, &items, historyRows+`SELECT * FROM history ORDER BY created_at DESC,id DESC LIMIT $5 OFFSET $6`, append(args, limit, (int64(page)-1)*int64(limit))...); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

type PaymentMethodSummary struct {
	ID        int       `db:"id" json:"id"`
	Provider  string    `db:"provider" json:"provider"`
	IsDefault bool      `db:"is_default" json:"is_default"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

func (s *CheckoutService) PaymentMethods(ctx context.Context, businessID string, customerID int) ([]PaymentMethodSummary, error) {
	var exists bool
	if err := s.db.GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM customers WHERE id=$1 AND business_id=$2)`, customerID, businessID); err != nil {
		return nil, err
	}
	if !exists {
		return nil, utils.ErrNotFound
	}
	items := []PaymentMethodSummary{}
	if err := s.db.SelectContext(ctx, &items, `SELECT pm.id,pm.provider,pm.is_default,pm.created_at FROM payment_methods pm
 JOIN customers c ON c.id=pm.customer_id WHERE pm.customer_id=$1 AND c.business_id=$2 ORDER BY pm.id`, customerID, businessID); err != nil {
		return nil, fmt.Errorf("list payment methods: %w", err)
	}
	return items, nil
}
