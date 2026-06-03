package repositories

import (
	"database/sql"
	"fmt"

	"pay-langgan/internal/database"
	"pay-langgan/internal/models"
)

type AddOnRepository struct {
	db *database.DB
}

func NewAddOnRepository(db *database.DB) *AddOnRepository {
	return &AddOnRepository{db: db}
}

func (r *AddOnRepository) FindAllByBusinessID(businessID string, page, limit int, search string) ([]models.AddOn, int, error) {
	countQuery := `SELECT COUNT(*) FROM add_ons a
		JOIN products p ON p.id = a.product_id
		JOIN services s ON s.id = p.service_id
		WHERE s.business_id = $1`
	args := []interface{}{businessID}
	argIdx := 2

	if search != "" {
		countQuery += fmt.Sprintf(" AND a.name ILIKE '%%%%' || $%d || '%%%%'", argIdx)
		args = append(args, search)
		argIdx++
	}

	var total int
	err := r.db.Get(&total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("count add-ons: %w", err)
	}

	offset := (page - 1) * limit
	dataQuery := `SELECT a.id, a.product_id, a.name, a.price, a.billing_cycle, a.meta, a.created_at
		FROM add_ons a
		JOIN products p ON p.id = a.product_id
		JOIN services s ON s.id = p.service_id
		WHERE s.business_id = $1`
	dataArgs := []interface{}{businessID}
	dataIdx := 2

	if search != "" {
		dataQuery += fmt.Sprintf(" AND a.name ILIKE '%%%%' || $%d || '%%%%'", dataIdx)
		dataArgs = append(dataArgs, search)
		dataIdx++
	}

	dataQuery += fmt.Sprintf(" ORDER BY a.id DESC LIMIT $%d OFFSET $%d", dataIdx, dataIdx+1)
	dataArgs = append(dataArgs, limit, offset)

	var addOns []models.AddOn
	err = r.db.Select(&addOns, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("find add-ons: %w", err)
	}

	return addOns, total, nil
}

func (r *AddOnRepository) FindByIDAndBusinessID(id int, businessID string) (*models.AddOn, error) {
	query := `SELECT a.id, a.product_id, a.name, a.price, a.billing_cycle, a.meta, a.created_at
		FROM add_ons a
		JOIN products p ON p.id = a.product_id
		JOIN services s ON s.id = p.service_id
		WHERE a.id = $1 AND s.business_id = $2`
	var addOn models.AddOn
	err := r.db.Get(&addOn, query, id, businessID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find add-on by id: %w", err)
	}
	return &addOn, nil
}

func (r *AddOnRepository) FindByID(id int) (*models.AddOn, error) {
	query := `SELECT id, product_id, name, price, billing_cycle, meta, created_at
		FROM add_ons WHERE id = $1`
	var addOn models.AddOn
	err := r.db.Get(&addOn, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find add-on by id: %w", err)
	}
	return &addOn, nil
}

func (r *AddOnRepository) Create(addOn *models.AddOn) error {
	if addOn.Meta == nil {
		addOn.Meta = make(models.JSONMap)
	}
	query := `INSERT INTO add_ons (product_id, name, price, billing_cycle, meta, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING id, created_at`
	err := r.db.QueryRow(query, addOn.ProductID, addOn.Name, addOn.Price, addOn.BillingCycle, addOn.Meta).
		Scan(&addOn.ID, &addOn.CreatedAt)
	if err != nil {
		return fmt.Errorf("create add-on: %w", err)
	}
	return nil
}

func (r *AddOnRepository) Update(addOn *models.AddOn) error {
	if addOn.Meta == nil {
		addOn.Meta = make(models.JSONMap)
	}
	query := `UPDATE add_ons SET name = $1, price = $2, billing_cycle = $3, meta = $4
		WHERE id = $5`
	_, err := r.db.Exec(query, addOn.Name, addOn.Price, addOn.BillingCycle, addOn.Meta, addOn.ID)
	if err != nil {
		return fmt.Errorf("update add-on: %w", err)
	}
	return nil
}

func (r *AddOnRepository) DeleteByIDAndBusinessID(id int, businessID string) error {
	query := `DELETE FROM add_ons a
		USING products p, services s
		WHERE a.product_id = p.id AND p.service_id = s.id AND a.id = $1 AND s.business_id = $2`
	_, err := r.db.Exec(query, id, businessID)
	if err != nil {
		return fmt.Errorf("delete add-on: %w", err)
	}
	return nil
}
