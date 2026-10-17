package customer

import (
	"database/sql"
	"fmt"

	"pay-langgan/internal/database"
	"pay-langgan/internal/models"
	"pay-langgan/internal/utils"
)

type CustomerRepository struct {
	db *database.DB
}

func NewCustomerRepository(db *database.DB) *CustomerRepository {
	return &CustomerRepository{db: db}
}

const customerStatusSQL = `CASE
 WHEN EXISTS (SELECT 1 FROM subscriptions s WHERE s.customer_id=c.id AND s.status='active') THEN 'active'
 WHEN EXISTS (SELECT 1 FROM subscriptions s WHERE s.customer_id=c.id AND s.status='trial') THEN 'trial'
 WHEN EXISTS (SELECT 1 FROM subscriptions s WHERE s.customer_id=c.id AND s.status='pending')
 OR NOT EXISTS (SELECT 1 FROM subscriptions s WHERE s.customer_id=c.id) THEN 'prospect'
 ELSE 'inactive' END`

func (r *CustomerRepository) FindAllByBusinessID(businessID string, page, limit int, search string, statuses ...string) ([]models.Customer, int, error) {
	where := ` FROM customers c WHERE c.business_id = $1`
	args := []interface{}{businessID}
	argIdx := 2

	if search != "" {
		where += fmt.Sprintf(" AND (c.name ILIKE '%%%%' || $%d || '%%%%' OR COALESCE(c.email, '') ILIKE '%%%%' || $%d || '%%%%')", argIdx, argIdx+1)
		args = append(args, search, search)
		argIdx += 2
	}
	if len(statuses) > 0 && statuses[0] != "" {
		where += fmt.Sprintf(" AND (%s)=$%d", customerStatusSQL, argIdx)
		args = append(args, statuses[0])
		argIdx++
	}

	var total int
	err := r.db.Get(&total, "SELECT COUNT(*)"+where, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("count customers: %w", err)
	}

	offset := (page - 1) * limit
	dataQuery := `SELECT c.id, c.business_id, c.name, c.email, c.contact, c.meta, c.created_at, ` + customerStatusSQL + ` AS status` + where
	dataArgs := append([]interface{}{}, args...)
	dataQuery += fmt.Sprintf(" ORDER BY c.id DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	dataArgs = append(dataArgs, limit, offset)

	var customers []models.Customer
	err = r.db.Select(&customers, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("find customers: %w", err)
	}

	return customers, total, nil
}

func (r *CustomerRepository) FindByIDAndBusinessID(id int, businessID string) (*models.Customer, error) {
	query := `SELECT id, business_id, name, email, contact, meta, created_at
		FROM customers WHERE id = $1 AND business_id = $2`
	var customer models.Customer
	err := r.db.Get(&customer, query, id, businessID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find customer by id: %w", err)
	}
	return &customer, nil
}

func (r *CustomerRepository) FindByID(id int) (*models.Customer, error) {
	query := `SELECT id, business_id, name, email, contact, meta, created_at
		FROM customers WHERE id = $1`
	var customer models.Customer
	err := r.db.Get(&customer, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find customer by id: %w", err)
	}
	return &customer, nil
}

func (r *CustomerRepository) Create(customer *models.Customer) error {
	if customer.Meta == nil {
		customer.Meta = make(models.JSONMap)
	}
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if customer.Email != nil && *customer.Email != "" {
		if _, err = tx.Exec(`SELECT pg_advisory_xact_lock(hashtext($1),hashtext(lower(btrim($2))))`, customer.BusinessID, *customer.Email); err != nil {
			return err
		}
		var exists bool
		if err = tx.Get(&exists, `SELECT EXISTS(SELECT 1 FROM customers WHERE business_id=$1 AND lower(btrim(email))=lower(btrim($2)))`, customer.BusinessID, *customer.Email); err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("%w: pelanggan dengan email ini sudah terdaftar", utils.ErrConflict)
		}
	}
	query := `INSERT INTO customers (business_id, name, email, contact, meta, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING id, created_at`
	err = tx.QueryRow(query, customer.BusinessID, customer.Name, customer.Email, customer.Contact, customer.Meta).
		Scan(&customer.ID, &customer.CreatedAt)
	if err != nil {
		return fmt.Errorf("create customer: %w", err)
	}
	return tx.Commit()
}

func (r *CustomerRepository) Update(customer *models.Customer) error {
	if customer.Meta == nil {
		customer.Meta = make(models.JSONMap)
	}
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if customer.Email != nil && *customer.Email != "" {
		if _, err = tx.Exec(`SELECT pg_advisory_xact_lock(hashtext($1),hashtext(lower(btrim($2))))`, customer.BusinessID, *customer.Email); err != nil {
			return err
		}
		var exists bool
		// Existing duplicate records remain editable when their email is unchanged.
		if err = tx.Get(&exists, `SELECT EXISTS(SELECT 1 FROM customers other WHERE other.business_id=$1
 AND other.id<>$3 AND lower(btrim(other.email))=lower(btrim($2)))
 AND EXISTS(SELECT 1 FROM customers own WHERE own.id=$3 AND own.business_id=$1
 AND lower(btrim(own.email)) IS DISTINCT FROM lower(btrim($2)))`, customer.BusinessID, *customer.Email, customer.ID); err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("%w: pelanggan dengan email ini sudah terdaftar", utils.ErrConflict)
		}
	}
	query := `UPDATE customers SET name = $1, email = $2, contact = $3, meta = $4
		WHERE id = $5 AND business_id = $6`
	_, err = tx.Exec(query, customer.Name, customer.Email, customer.Contact, customer.Meta, customer.ID, customer.BusinessID)
	if err != nil {
		return fmt.Errorf("update customer: %w", err)
	}
	return tx.Commit()
}

func (r *CustomerRepository) DeleteByIDAndBusinessID(id int, businessID string) error {
	query := `DELETE FROM customers WHERE id = $1 AND business_id = $2`
	_, err := r.db.Exec(query, id, businessID)
	if err != nil {
		return fmt.Errorf("delete customer: %w", err)
	}
	return nil
}
