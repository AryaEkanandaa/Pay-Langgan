package repositories

import (
	"database/sql"
	"fmt"

	"pay-langgan/internal/database"
	"pay-langgan/internal/models"
)

type ProductRepository struct {
	db *database.DB
}

func NewProductRepository(db *database.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) FindAllByBusinessID(businessID string, page, limit int, search string) ([]models.Product, int, error) {
	countQuery := `SELECT COUNT(*) FROM products p
		JOIN services s ON s.id = p.service_id
		WHERE s.business_id = $1`
	args := []interface{}{businessID}
	argIdx := 2

	if search != "" {
		countQuery += fmt.Sprintf(" AND p.name ILIKE '%%%%' || $%d || '%%%%'", argIdx)
		args = append(args, search)
		argIdx++
	}

	var total int
	err := r.db.Get(&total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("count products: %w", err)
	}

	offset := (page - 1) * limit
	dataQuery := `SELECT p.id, p.service_id, p.name, p.description, p.status, p.meta, p.created_at
		FROM products p
		JOIN services s ON s.id = p.service_id
		WHERE s.business_id = $1`
	dataArgs := []interface{}{businessID}
	dataIdx := 2

	if search != "" {
		dataQuery += fmt.Sprintf(" AND p.name ILIKE '%%%%' || $%d || '%%%%'", dataIdx)
		dataArgs = append(dataArgs, search)
		dataIdx++
	}

	dataQuery += fmt.Sprintf(" ORDER BY p.id DESC LIMIT $%d OFFSET $%d", dataIdx, dataIdx+1)
	dataArgs = append(dataArgs, limit, offset)

	var products []models.Product
	err = r.db.Select(&products, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("find products: %w", err)
	}

	return products, total, nil
}

func (r *ProductRepository) FindByIDAndBusinessID(id int, businessID string) (*models.Product, error) {
	query := `SELECT p.id, p.service_id, p.name, p.description, p.status, p.meta, p.created_at
		FROM products p
		JOIN services s ON s.id = p.service_id
		WHERE p.id = $1 AND s.business_id = $2`
	var product models.Product
	err := r.db.Get(&product, query, id, businessID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find product by id: %w", err)
	}
	return &product, nil
}

func (r *ProductRepository) FindByID(id int) (*models.Product, error) {
	query := `SELECT id, service_id, name, description, status, meta, created_at
		FROM products WHERE id = $1`
	var product models.Product
	err := r.db.Get(&product, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find product by id: %w", err)
	}
	return &product, nil
}

func (r *ProductRepository) Create(product *models.Product) error {
	if product.Meta == nil {
		product.Meta = make(models.JSONMap)
	}
	if product.Status == "" {
		product.Status = "active"
	}
	query := `INSERT INTO products (service_id, name, description, status, meta, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING id, created_at`
	err := r.db.QueryRow(query, product.ServiceID, product.Name, product.Description, product.Status, product.Meta).
		Scan(&product.ID, &product.CreatedAt)
	if err != nil {
		return fmt.Errorf("create product: %w", err)
	}
	return nil
}

func (r *ProductRepository) Update(product *models.Product) error {
	if product.Meta == nil {
		product.Meta = make(models.JSONMap)
	}
	query := `UPDATE products SET name = $1, description = $2, status = $3, meta = $4
		WHERE id = $5`
	_, err := r.db.Exec(query, product.Name, product.Description, product.Status, product.Meta, product.ID)
	if err != nil {
		return fmt.Errorf("update product: %w", err)
	}
	return nil
}

func (r *ProductRepository) DeleteByIDAndBusinessID(id int, businessID string) error {
	query := `DELETE FROM products p
		USING services s
		WHERE p.service_id = s.id AND p.id = $1 AND s.business_id = $2`
	_, err := r.db.Exec(query, id, businessID)
	if err != nil {
		return fmt.Errorf("delete product: %w", err)
	}
	return nil
}
