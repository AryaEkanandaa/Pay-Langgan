package customer

import (
	"database/sql"
	"fmt"

	"pay-langgan/internal/database"
	"pay-langgan/internal/models"
)

type CustomerRepository struct {
	db *database.DB
}

func NewCustomerRepository(db *database.DB) *CustomerRepository {
	return &CustomerRepository{db: db}
}

func (r *CustomerRepository) FindAllByBusinessID(businessID string, page, limit int, search string) ([]models.Customer, int, error) {
	countQuery := `SELECT COUNT(*) FROM customers WHERE business_id = $1`
	args := []interface{}{businessID}
	argIdx := 2

	if search != "" {
		countQuery += fmt.Sprintf(" AND (name ILIKE '%%%%' || $%d || '%%%%' OR COALESCE(email, '') ILIKE '%%%%' || $%d || '%%%%')", argIdx, argIdx+1)
		args = append(args, search, search)
		argIdx += 2
	}

	var total int
	err := r.db.Get(&total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("count customers: %w", err)
	}

	offset := (page - 1) * limit
	dataQuery := `SELECT id, business_id, name, email, contact, meta, created_at
		FROM customers WHERE business_id = $1`
	dataArgs := []interface{}{businessID}
	dataIdx := 2

	if search != "" {
		dataQuery += fmt.Sprintf(" AND (name ILIKE '%%%%' || $%d || '%%%%' OR COALESCE(email, '') ILIKE '%%%%' || $%d || '%%%%')", dataIdx, dataIdx+1)
		dataArgs = append(dataArgs, search, search)
		dataIdx += 2
	}

	dataQuery += fmt.Sprintf(" ORDER BY id DESC LIMIT $%d OFFSET $%d", dataIdx, dataIdx+1)
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
	query := `INSERT INTO customers (business_id, name, email, contact, meta, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING id, created_at`
	err := r.db.QueryRow(query, customer.BusinessID, customer.Name, customer.Email, customer.Contact, customer.Meta).
		Scan(&customer.ID, &customer.CreatedAt)
	if err != nil {
		return fmt.Errorf("create customer: %w", err)
	}
	return nil
}

func (r *CustomerRepository) Update(customer *models.Customer) error {
	if customer.Meta == nil {
		customer.Meta = make(models.JSONMap)
	}
	query := `UPDATE customers SET name = $1, email = $2, contact = $3, meta = $4
		WHERE id = $5 AND business_id = $6`
	_, err := r.db.Exec(query, customer.Name, customer.Email, customer.Contact, customer.Meta, customer.ID, customer.BusinessID)
	if err != nil {
		return fmt.Errorf("update customer: %w", err)
	}
	return nil
}

func (r *CustomerRepository) DeleteByIDAndBusinessID(id int, businessID string) error {
	query := `DELETE FROM customers WHERE id = $1 AND business_id = $2`
	_, err := r.db.Exec(query, id, businessID)
	if err != nil {
		return fmt.Errorf("delete customer: %w", err)
	}
	return nil
}
