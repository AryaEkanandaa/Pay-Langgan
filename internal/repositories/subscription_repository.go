package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"pay-langgan/internal/database"
	"pay-langgan/internal/models"

	"github.com/jmoiron/sqlx"
)

type SubscriptionRepository struct {
	db *database.DB
}

func NewSubscriptionRepository(db *database.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

func (r *SubscriptionRepository) FindAllByBusinessID(businessID string, page, limit int, status, search string) ([]models.Subscription, int, error) {
	baseWhere := "FROM subscriptions s JOIN customers c ON c.id = s.customer_id WHERE c.business_id = $1"
	args := []interface{}{businessID}
	argIdx := 2

	if status != "" {
		baseWhere += fmt.Sprintf(" AND s.status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}

	if search != "" {
		baseWhere += fmt.Sprintf(" AND (c.name ILIKE '%%%%' || $%d || '%%%%' OR COALESCE(c.email, '') ILIKE '%%%%' || $%d || '%%%%')", argIdx, argIdx+1)
		args = append(args, search, search)
		argIdx += 2
	}

	countQuery := "SELECT COUNT(*) " + baseWhere
	var total int
	err := r.db.Get(&total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("count subscriptions: %w", err)
	}

	offset := (page - 1) * limit
	dataQuery := fmt.Sprintf("SELECT s.id, s.customer_id, s.plan_id, s.status, s.start_date, s.next_billing_date, s.end_date, s.trial_ends_at, s.meta, s.created_at %s ORDER BY s.id DESC LIMIT $%d OFFSET $%d", baseWhere, argIdx, argIdx+1)
	dataArgs := append(args, limit, offset)

	var subscriptions []models.Subscription
	err = r.db.Select(&subscriptions, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("find subscriptions: %w", err)
	}

	return subscriptions, total, nil
}

func (r *SubscriptionRepository) FindByIDAndBusinessID(id int, businessID string) (*models.Subscription, error) {
	query := `SELECT s.id, s.customer_id, s.plan_id, s.status, s.start_date, s.next_billing_date, s.end_date, s.trial_ends_at, s.meta, s.created_at
		FROM subscriptions s
		JOIN customers c ON c.id = s.customer_id
		WHERE s.id = $1 AND c.business_id = $2`
	var sub models.Subscription
	err := r.db.Get(&sub, query, id, businessID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find subscription by id: %w", err)
	}
	return &sub, nil
}

func (r *SubscriptionRepository) FindByID(id int) (*models.Subscription, error) {
	query := `SELECT id, customer_id, plan_id, status, start_date, next_billing_date, end_date, trial_ends_at, meta, created_at
		FROM subscriptions WHERE id = $1`
	var sub models.Subscription
	err := r.db.Get(&sub, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find subscription by id: %w", err)
	}
	return &sub, nil
}

func (r *SubscriptionRepository) Create(tx *sqlx.Tx, sub *models.Subscription) error {
	if sub.Meta == nil {
		sub.Meta = make(models.JSONMap)
	}
	query := `INSERT INTO subscriptions (customer_id, plan_id, status, start_date, next_billing_date, end_date, trial_ends_at, meta, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		RETURNING id, created_at`
	err := tx.QueryRow(query, sub.CustomerID, sub.PlanID, sub.Status, sub.StartDate, sub.NextBillingDate, sub.EndDate, sub.TrialEndsAt, sub.Meta).
		Scan(&sub.ID, &sub.CreatedAt)
	if err != nil {
		return fmt.Errorf("create subscription: %w", err)
	}
	return nil
}

func (r *SubscriptionRepository) UpdateStatus(tx *sqlx.Tx, sub *models.Subscription) error {
	if sub.Meta == nil {
		sub.Meta = make(models.JSONMap)
	}
	query := `UPDATE subscriptions SET status = $1, end_date = $2, next_billing_date = $3, trial_ends_at = $4, meta = $5
		WHERE id = $6`
	_, err := tx.Exec(query, sub.Status, sub.EndDate, sub.NextBillingDate, sub.TrialEndsAt, sub.Meta, sub.ID)
	if err != nil {
		return fmt.Errorf("update subscription status: %w", err)
	}
	return nil
}

func timePtr(t time.Time) *time.Time {
	return &t
}
