package subscription

import (
	"fmt"

	"pay-langgan/internal/database"
	"pay-langgan/internal/models"

	"github.com/jmoiron/sqlx"
)

type SubscriptionAddOnRepository struct {
	db *database.DB
}

func NewSubscriptionAddOnRepository(db *database.DB) *SubscriptionAddOnRepository {
	return &SubscriptionAddOnRepository{db: db}
}

func (r *SubscriptionAddOnRepository) Upsert(tx *sqlx.Tx, item *models.SubscriptionAddOn) error {
	query := `UPDATE subscription_add_ons SET quantity = $1
		WHERE subscription_id = $2 AND add_on_id = $3`
	res, err := tx.Exec(query, item.Quantity, item.SubscriptionID, item.AddOnID)
	if err != nil {
		return fmt.Errorf("upsert subscription add-on update: %w", err)
	}

	rows, _ := res.RowsAffected()
	if rows > 0 {
		return nil
	}

	query = `INSERT INTO subscription_add_ons (subscription_id, add_on_id, quantity, created_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING id, created_at`
	err = tx.QueryRow(query, item.SubscriptionID, item.AddOnID, item.Quantity).
		Scan(&item.ID, &item.CreatedAt)
	if err != nil {
		return fmt.Errorf("upsert subscription add-on insert: %w", err)
	}
	return nil
}

func (r *SubscriptionAddOnRepository) DeleteBySubscriptionIDAndAddOnID(tx *sqlx.Tx, subscriptionID, addOnID int) error {
	query := `DELETE FROM subscription_add_ons WHERE subscription_id = $1 AND add_on_id = $2`
	_, err := tx.Exec(query, subscriptionID, addOnID)
	if err != nil {
		return fmt.Errorf("delete subscription add-on: %w", err)
	}
	return nil
}

func (r *SubscriptionAddOnRepository) FindBySubscriptionID(subscriptionID int) ([]models.SubscriptionAddOn, error) {
	query := `SELECT id, subscription_id, add_on_id, quantity, created_at
		FROM subscription_add_ons WHERE subscription_id = $1`
	var items []models.SubscriptionAddOn
	err := r.db.Select(&items, query, subscriptionID)
	if err != nil {
		return nil, fmt.Errorf("find subscription add-ons: %w", err)
	}
	return items, nil
}
