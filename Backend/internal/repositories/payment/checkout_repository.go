package payment

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"pay-langgan/internal/database"
	"pay-langgan/internal/models"
)

type CheckoutRepository struct {
	db *database.DB
}

func NewCheckoutRepository(db *database.DB) *CheckoutRepository {
	return &CheckoutRepository{db: db}
}

const checkoutRecordColumns = `SELECT cs.id AS checkout_id, cs.invoice_id, cs.payment_id,
 cs.reference_id, cs.provider_session_id, cs.payment_link_url, cs.status AS checkout_status,
 cs.expires_at,
 i.invoice_number, i.amount AS invoice_amount, i.status AS invoice_status,
 c.id AS customer_id, c.name AS customer_name, c.email AS default_email,
 c.contact AS default_mobile, b.id AS business_id, b.name AS business_name, pl.name AS plan_name
 FROM checkout_sessions cs
 JOIN invoices i ON i.id = cs.invoice_id
 JOIN subscriptions s ON s.id = i.subscription_id
 JOIN customers c ON c.id = s.customer_id
 JOIN businesses b ON b.id = c.business_id
 JOIN plans pl ON pl.id = s.plan_id`

func (r *CheckoutRepository) FindInvoiceForCheckout(tx *sqlx.Tx, invoiceID int, businessID string) (*models.CheckoutRecord, error) {
	var record models.CheckoutRecord
	err := tx.Get(&record, `SELECT i.id AS invoice_id, i.invoice_number, i.amount AS invoice_amount,
 i.status AS invoice_status, c.id AS customer_id, c.name AS customer_name,
 c.email AS default_email, c.contact AS default_mobile, b.name AS business_name,
 pl.name AS plan_name
 FROM invoices i
 JOIN subscriptions s ON s.id = i.subscription_id
 JOIN customers c ON c.id = s.customer_id
 JOIN businesses b ON b.id = c.business_id
 JOIN plans pl ON pl.id = s.plan_id
 WHERE i.id = $1 AND c.business_id = $2 FOR UPDATE OF i`, invoiceID, businessID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find invoice for checkout: %w", err)
	}
	return &record, nil
}

func (r *CheckoutRepository) Create(tx *sqlx.Tx, checkout *models.CheckoutSession) error {
	return tx.QueryRow(`INSERT INTO checkout_sessions
 (invoice_id, public_token_hash, reference_id, provider, status, expires_at)
 VALUES ($1, $2, $3, 'xendit', 'created', $4)
 RETURNING id, created_at, updated_at`, checkout.InvoiceID, checkout.PublicTokenHash,
		checkout.ReferenceID, checkout.ExpiresAt).
		Scan(&checkout.ID, &checkout.CreatedAt, &checkout.UpdatedAt)
}

func (r *CheckoutRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*models.CheckoutRecord, error) {
	var record models.CheckoutRecord
	err := r.db.GetContext(ctx, &record, checkoutRecordColumns+` WHERE cs.public_token_hash = $1`, tokenHash)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find checkout: %w", err)
	}
	return &record, nil
}

func (r *CheckoutRepository) FindByID(ctx context.Context, id int64) (*models.CheckoutRecord, error) {
	var record models.CheckoutRecord
	err := r.db.GetContext(ctx, &record, checkoutRecordColumns+` WHERE cs.id=$1`, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *CheckoutRepository) Claim(ctx context.Context, checkoutID int64, email, mobile string) (bool, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin claim checkout: %w", err)
	}
	defer tx.Rollback()
	var invoiceID int
	err = tx.GetContext(ctx, &invoiceID, `SELECT cs.invoice_id FROM checkout_sessions cs
 JOIN invoices i ON i.id = cs.invoice_id
 WHERE cs.id = $1 AND cs.status = 'created' AND cs.expires_at > (NOW() AT TIME ZONE 'UTC')
 AND i.status = 'pending' FOR UPDATE OF i`, checkoutID)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("lock invoice for checkout: %w", err)
	}
	var active int
	if err := tx.GetContext(ctx, &active, `SELECT count(*) FROM checkout_sessions
 WHERE invoice_id = $1 AND id <> $2 AND status IN ('creating', 'active')`, invoiceID, checkoutID); err != nil {
		return false, fmt.Errorf("check competing checkout: %w", err)
	}
	if active > 0 {
		return false, nil
	}
	if err := tx.GetContext(ctx, &active, `SELECT count(*) FROM autopay_charges WHERE invoice_id=$1 AND status IN ('creating','pending','review_required')`, invoiceID); err != nil {
		return false, err
	}
	if active > 0 {
		return false, nil
	}
	result, err := tx.ExecContext(ctx, `UPDATE checkout_sessions
 SET status = 'creating', customer_email = $2, customer_mobile = $3, updated_at = (NOW() AT TIME ZONE 'UTC')
 WHERE id = $1 AND status = 'created'`, checkoutID, email, mobile)
	if err != nil {
		return false, fmt.Errorf("claim checkout: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil || rows != 1 {
		return false, err
	}
	_, err = tx.ExecContext(ctx, `UPDATE checkout_sessions SET status = 'expired',
 updated_at = (NOW() AT TIME ZONE 'UTC')
 WHERE invoice_id = $1 AND id <> $2 AND status IN ('created', 'failed')`, invoiceID, checkoutID)
	if err != nil {
		return false, fmt.Errorf("expire competing checkout links: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit checkout claim: %w", err)
	}
	return true, nil
}

func (r *CheckoutRepository) MarkFailed(ctx context.Context, checkoutID int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE checkout_sessions SET status = 'failed',last_error='session_creation_rejected',review_required=FALSE, updated_at = (NOW() AT TIME ZONE 'UTC')
 WHERE id = $1 AND status = 'creating'`, checkoutID)
	return err
}

// Persist the provider ID before activating the payment. This permits recovery
// if the following transaction fails. Never store the Components SDK key.
func (r *CheckoutRepository) RememberSession(ctx context.Context, checkoutID int64, sessionID string) error {
	result, err := r.db.ExecContext(ctx, `UPDATE checkout_sessions SET provider_session_id=$2
 WHERE id=$1 AND status IN ('creating','active','completed','expired')
 AND (provider_session_id IS NULL OR provider_session_id=$2)`, checkoutID, sessionID)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return fmt.Errorf("checkout session identity conflict")
	}
	return nil
}

func (r *CheckoutRepository) Activate(ctx context.Context, checkoutID int64, invoiceID int, amount float64, providerSessionID string, expiresAt time.Time) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin activate checkout: %w", err)
	}
	defer tx.Rollback()
	var current struct {
		Status    string  `db:"status"`
		SessionID *string `db:"provider_session_id"`
	}
	// Use the same invoice -> checkout lock order as claim and reconciliation.
	if _, err := tx.ExecContext(ctx, `SELECT id FROM invoices WHERE id=$1 FOR UPDATE`, invoiceID); err != nil {
		return err
	}
	if err := tx.GetContext(ctx, &current, `SELECT status, provider_session_id FROM checkout_sessions WHERE id=$1 FOR UPDATE`, checkoutID); err != nil {
		return err
	}
	if current.SessionID != nil && *current.SessionID != providerSessionID {
		return fmt.Errorf("checkout session identity conflict")
	}
	if current.Status != "creating" {
		if current.SessionID != nil && *current.SessionID == providerSessionID {
			return tx.Commit()
		}
		return fmt.Errorf("checkout activation lost its claim")
	}
	var paymentID int
	err = tx.QueryRowContext(ctx, `INSERT INTO payments (invoice_id, provider, provider_transaction_id, amount, status, created_at)
 VALUES ($1, 'xendit', $2, $3, 'pending', (NOW() AT TIME ZONE 'UTC')) RETURNING id`, invoiceID, providerSessionID, amount).Scan(&paymentID)
	if err != nil {
		return fmt.Errorf("create pending payment: %w", err)
	}
	result, err := tx.ExecContext(ctx, `UPDATE checkout_sessions SET payment_id = $2,
 provider_session_id = $3, payment_link_url = NULL, expires_at = LEAST(expires_at, $4), status = 'active', updated_at = (NOW() AT TIME ZONE 'UTC')
 WHERE id = $1 AND status = 'creating'`, checkoutID, paymentID, providerSessionID, expiresAt)
	if err != nil {
		return fmt.Errorf("activate checkout: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil || rows != 1 {
		return fmt.Errorf("checkout activation lost its claim")
	}
	return tx.Commit()
}
