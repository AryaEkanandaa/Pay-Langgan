package catalog

import (
	"database/sql"
	"fmt"

	"pay-langgan/internal/database"
	"pay-langgan/internal/models"
)

type ServiceRepository struct {
	db *database.DB
}

func NewServiceRepository(db *database.DB) *ServiceRepository {
	return &ServiceRepository{db: db}
}

func (r *ServiceRepository) FindAllByBusinessID(businessID string, page, limit int, search string) ([]models.Service, int, error) {
	countQuery := `SELECT COUNT(*) FROM services WHERE business_id = $1`
	args := []interface{}{businessID}
	argIdx := 2

	if search != "" {
		countQuery += fmt.Sprintf(" AND name ILIKE '%%%%' || $%d || '%%%%'", argIdx)
		args = append(args, search)
		argIdx++
	}

	var total int
	err := r.db.Get(&total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("count services: %w", err)
	}

	offset := (page - 1) * limit
	dataQuery := `SELECT id, business_id, name, description, meta, created_at
		FROM services WHERE business_id = $1`
	dataArgs := []interface{}{businessID}
	dataIdx := 2

	if search != "" {
		dataQuery += fmt.Sprintf(" AND name ILIKE '%%%%' || $%d || '%%%%'", dataIdx)
		dataArgs = append(dataArgs, search)
		dataIdx++
	}

	dataQuery += fmt.Sprintf(" ORDER BY id DESC LIMIT $%d OFFSET $%d", dataIdx, dataIdx+1)
	dataArgs = append(dataArgs, limit, offset)

	var services []models.Service
	err = r.db.Select(&services, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("find services: %w", err)
	}

	return services, total, nil
}

func (r *ServiceRepository) FindByIDAndBusinessID(id int, businessID string) (*models.Service, error) {
	query := `SELECT id, business_id, name, description, meta, created_at
		FROM services WHERE id = $1 AND business_id = $2`
	var service models.Service
	err := r.db.Get(&service, query, id, businessID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find service by id: %w", err)
	}
	return &service, nil
}

func (r *ServiceRepository) FindByID(id int) (*models.Service, error) {
	query := `SELECT id, business_id, name, description, meta, created_at
		FROM services WHERE id = $1`
	var service models.Service
	err := r.db.Get(&service, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find service by id: %w", err)
	}
	return &service, nil
}

func (r *ServiceRepository) Create(service *models.Service) error {
	if service.Meta == nil {
		service.Meta = make(models.JSONMap)
	}
	query := `INSERT INTO services (business_id, name, description, meta, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id, created_at`
	err := r.db.QueryRow(query, service.BusinessID, service.Name, service.Description, service.Meta).
		Scan(&service.ID, &service.CreatedAt)
	if err != nil {
		return fmt.Errorf("create service: %w", err)
	}
	return nil
}

func (r *ServiceRepository) Update(service *models.Service) error {
	if service.Meta == nil {
		service.Meta = make(models.JSONMap)
	}
	query := `UPDATE services SET name = $1, description = $2, meta = $3
		WHERE id = $4 AND business_id = $5`
	_, err := r.db.Exec(query, service.Name, service.Description, service.Meta, service.ID, service.BusinessID)
	if err != nil {
		return fmt.Errorf("update service: %w", err)
	}
	return nil
}

func (r *ServiceRepository) DeleteByIDAndBusinessID(id int, businessID string) error {
	query := `DELETE FROM services WHERE id = $1 AND business_id = $2`
	_, err := r.db.Exec(query, id, businessID)
	if err != nil {
		return fmt.Errorf("delete service: %w", err)
	}
	return nil
}
