package identity

import (
	"database/sql"
	"fmt"

	"pay-langgan/internal/database"
	"pay-langgan/internal/models"
)

type UserRepository struct {
	db *database.DB
}

func NewUserRepository(db *database.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	query := `SELECT id, business_id, name, email, password, role, created_at, updated_at
		FROM users WHERE email = $1`
	var user models.User
	if err := r.db.Get(&user, query, email); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) FindAllByBusinessID(businessID string) ([]models.User, error) {
	query := `SELECT id, business_id, name, email, password, role, created_at, updated_at
		FROM users WHERE business_id = $1 ORDER BY id ASC`
	var users []models.User
	if err := r.db.Select(&users, query, businessID); err != nil {
		return nil, fmt.Errorf("find users by business: %w", err)
	}
	return users, nil
}

func (r *UserRepository) Create(user *models.User) error {
	query := `INSERT INTO users (business_id, name, email, password, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, created_at, updated_at`
	if err := r.db.QueryRow(query, user.BusinessID, user.Name, user.Email, user.Password, user.Role).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt); err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}
