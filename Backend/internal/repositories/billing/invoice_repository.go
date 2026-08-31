package billing

import (
	"database/sql"
	"fmt"

	"pay-langgan/internal/database"
	"pay-langgan/internal/models"

	"github.com/jmoiron/sqlx"
)

type InvoiceRepository struct {
	db *database.DB
}

func NewInvoiceRepository(db *database.DB) *InvoiceRepository {
	return &InvoiceRepository{db: db}
}

func (r *InvoiceRepository) Create(tx *sqlx.Tx, invoice *models.Invoice) error {
	query := `INSERT INTO invoices (subscription_id, invoice_number, amount, status, due_date, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING id, created_at`
	if err := tx.QueryRow(query, invoice.SubscriptionID, invoice.InvoiceNumber, invoice.Amount, invoice.Status, invoice.DueDate).
		Scan(&invoice.ID, &invoice.CreatedAt); err != nil {
		return fmt.Errorf("create invoice: %w", err)
	}
	return nil
}

const invoiceFromBusiness = `
	FROM invoices i
	JOIN subscriptions s ON s.id = i.subscription_id
	JOIN customers c ON c.id = s.customer_id
	JOIN plans pl ON pl.id = s.plan_id
	WHERE c.business_id = $1`

func (r *InvoiceRepository) FindAllByBusinessID(businessID string, page, limit int, status, search string) ([]models.InvoiceListItem, int, error) {
	baseWhere := invoiceFromBusiness
	args := []interface{}{businessID}
	argIndex := 2

	if status != "" {
		baseWhere += fmt.Sprintf(" AND i.status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}
	if search != "" {
		baseWhere += fmt.Sprintf(" AND (i.invoice_number ILIKE '%%' || $%d || '%%' OR c.name ILIKE '%%' || $%d || '%%' OR COALESCE(c.email, '') ILIKE '%%' || $%d || '%%')", argIndex, argIndex, argIndex)
		args = append(args, search)
		argIndex++
	}

	var total int
	if err := r.db.Get(&total, "SELECT COUNT(*) "+baseWhere, args...); err != nil {
		return nil, 0, fmt.Errorf("count invoices: %w", err)
	}

	offset := (page - 1) * limit
	query := `SELECT i.id, i.subscription_id, i.invoice_number, i.amount, i.status, i.due_date, i.paid_at, i.created_at,
		c.id AS customer_id, c.name AS customer_name, c.email AS customer_email, pl.name AS plan_name ` + baseWhere +
		fmt.Sprintf(" ORDER BY i.created_at DESC, i.id DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	queryArgs := append(args, limit, offset)

	var invoices []models.InvoiceListItem
	if err := r.db.Select(&invoices, query, queryArgs...); err != nil {
		return nil, 0, fmt.Errorf("find invoices: %w", err)
	}
	return invoices, total, nil
}

func (r *InvoiceRepository) FindByIDAndBusinessID(id int, businessID string) (*models.InvoiceListItem, error) {
	query := `SELECT i.id, i.subscription_id, i.invoice_number, i.amount, i.status, i.due_date, i.paid_at, i.created_at,
		c.id AS customer_id, c.name AS customer_name, c.email AS customer_email, pl.name AS plan_name ` + invoiceFromBusiness + ` AND i.id = $2`
	var invoice models.InvoiceListItem
	if err := r.db.Get(&invoice, query, businessID, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find invoice by id: %w", err)
	}
	return &invoice, nil
}

func (r *InvoiceRepository) FindItems(invoiceID int) ([]models.InvoiceItemResponse, error) {
	query := `SELECT id, invoice_id, item_type, description, quantity, unit_price, subtotal, meta, created_at
		FROM invoice_items WHERE invoice_id = $1 ORDER BY id ASC`
	var items []models.InvoiceItemResponse
	if err := r.db.Select(&items, query, invoiceID); err != nil {
		return nil, fmt.Errorf("find invoice items: %w", err)
	}
	return items, nil
}

func (r *InvoiceRepository) FindLatestPayment(invoiceID int) (*models.Payment, error) {
	query := `SELECT id, invoice_id, provider, provider_transaction_id, amount, status, paid_at, created_at
		FROM payments WHERE invoice_id = $1 ORDER BY created_at DESC, id DESC LIMIT 1`
	var payment models.Payment
	if err := r.db.Get(&payment, query, invoiceID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find invoice payment: %w", err)
	}
	return &payment, nil
}

func (r *InvoiceRepository) MarkPaid(tx *sqlx.Tx, id int, businessID string) (float64, bool, error) {
	query := `UPDATE invoices i
		SET status = 'paid', paid_at = NOW()
		FROM subscriptions s
		JOIN customers c ON c.id = s.customer_id
		WHERE i.subscription_id = s.id
		  AND i.id = $1
		  AND c.business_id = $2
		  AND i.status NOT IN ('paid', 'cancelled')
		RETURNING i.amount`
	var amount float64
	if err := tx.QueryRow(query, id, businessID).Scan(&amount); err != nil {
		if err == sql.ErrNoRows {
			return 0, false, nil
		}
		return 0, false, fmt.Errorf("mark invoice paid: %w", err)
	}

	_, err := tx.Exec(`INSERT INTO payments (invoice_id, provider, amount, status, paid_at, created_at)
		VALUES ($1, 'manual', $2, 'success', NOW(), NOW())`, id, amount)
	if err != nil {
		return 0, false, fmt.Errorf("record manual payment: %w", err)
	}
	return amount, true, nil
}
