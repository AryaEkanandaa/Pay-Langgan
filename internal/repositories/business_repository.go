package repositories

import (
	"database/sql"
	"fmt"

	"pay-langgan/internal/database"
	"pay-langgan/internal/models"
)

type BusinessRepository struct {
	db *database.DB
}

func NewBusinessRepository(db *database.DB) *BusinessRepository {
	return &BusinessRepository{db: db}
}

func (r *BusinessRepository) FindByID(id string) (*models.Business, error) {
	query := `SELECT id, name, status, deleted, created_at, updated_at, meta
		FROM businesses WHERE id = $1 AND deleted = FALSE`
	var business models.Business
	err := r.db.Get(&business, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find business by id: %w", err)
	}
	return &business, nil
}

func (r *BusinessRepository) Update(business *models.Business) error {
	query := `UPDATE businesses SET name = $1, meta = $2, updated_at = NOW()
		WHERE id = $3 AND deleted = FALSE`
	_, err := r.db.Exec(query, business.Name, business.Meta, business.ID)
	if err != nil {
		return fmt.Errorf("update business: %w", err)
	}
	return nil
}
