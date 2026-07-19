package identity

import (
	"database/sql"
	"fmt"

	"pay-langgan/internal/database"
	"pay-langgan/internal/models"
)

type AuthRepository struct {
	db *database.DB
}

func NewAuthRepository(db *database.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) FindUserByEmail(email string) (*models.User, error) {
	query := `SELECT id, business_id, name, email, password, role, created_at, updated_at
		FROM users WHERE email = $1`

	var user models.User
	err := r.db.Get(&user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	return &user, nil
}

func (r *AuthRepository) FindBusinessByID(id string) (*models.Business, error) {
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

func (r *AuthRepository) CreateBusinessAndUserTx(business *models.Business, user *models.User) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	bizQuery := `INSERT INTO businesses (id, name, status, deleted, created_at, updated_at, meta)
		VALUES ($1, $2, 'active', FALSE, NOW(), NOW(), '{}'::jsonb)`

	_, err = tx.Exec(bizQuery, business.ID, business.Name)
	if err != nil {
		return fmt.Errorf("insert business: %w", err)
	}

	userQuery := `INSERT INTO users (business_id, name, email, password, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 'admin', NOW(), NOW())
		RETURNING id`

	err = tx.QueryRow(userQuery, business.ID, user.Name, user.Email, user.Password).Scan(&user.ID)
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}

	return tx.Commit()
}
