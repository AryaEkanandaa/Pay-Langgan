package catalog

import (
	"database/sql"
	"fmt"

	"pay-langgan/internal/database"
	"pay-langgan/internal/models"
)

type PlanRepository struct {
	db *database.DB
}

func NewPlanRepository(db *database.DB) *PlanRepository {
	return &PlanRepository{db: db}
}

func (r *PlanRepository) FindAllByBusinessID(businessID string, page, limit int, search string) ([]models.Plan, int, error) {
	countQuery := `SELECT COUNT(*) FROM plans pl
		JOIN products p ON p.id = pl.product_id
		JOIN services s ON s.id = p.service_id
		WHERE s.business_id = $1`
	args := []interface{}{businessID}
	argIdx := 2

	if search != "" {
		countQuery += fmt.Sprintf(" AND pl.name ILIKE '%%%%' || $%d || '%%%%'", argIdx)
		args = append(args, search)
		argIdx++
	}

	var total int
	err := r.db.Get(&total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("count plans: %w", err)
	}

	offset := (page - 1) * limit
	dataQuery := `SELECT pl.id, pl.product_id, pl.name, pl.price, pl.billing_cycle, pl.trial_days, pl.meta, pl.created_at
		FROM plans pl
		JOIN products p ON p.id = pl.product_id
		JOIN services s ON s.id = p.service_id
		WHERE s.business_id = $1`
	dataArgs := []interface{}{businessID}
	dataIdx := 2

	if search != "" {
		dataQuery += fmt.Sprintf(" AND pl.name ILIKE '%%%%' || $%d || '%%%%'", dataIdx)
		dataArgs = append(dataArgs, search)
		dataIdx++
	}

	dataQuery += fmt.Sprintf(" ORDER BY pl.id DESC LIMIT $%d OFFSET $%d", dataIdx, dataIdx+1)
	dataArgs = append(dataArgs, limit, offset)

	var plans []models.Plan
	err = r.db.Select(&plans, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("find plans: %w", err)
	}

	return plans, total, nil
}

func (r *PlanRepository) FindByIDAndBusinessID(id int, businessID string) (*models.Plan, error) {
	query := `SELECT pl.id, pl.product_id, pl.name, pl.price, pl.billing_cycle, pl.trial_days, pl.meta, pl.created_at
		FROM plans pl
		JOIN products p ON p.id = pl.product_id
		JOIN services s ON s.id = p.service_id
		WHERE pl.id = $1 AND s.business_id = $2`
	var plan models.Plan
	err := r.db.Get(&plan, query, id, businessID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find plan by id: %w", err)
	}
	return &plan, nil
}

func (r *PlanRepository) FindByID(id int) (*models.Plan, error) {
	query := `SELECT id, product_id, name, price, billing_cycle, trial_days, meta, created_at
		FROM plans WHERE id = $1`
	var plan models.Plan
	err := r.db.Get(&plan, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find plan by id: %w", err)
	}
	return &plan, nil
}

func (r *PlanRepository) Create(plan *models.Plan) error {
	if plan.Meta == nil {
		plan.Meta = make(models.JSONMap)
	}
	query := `INSERT INTO plans (product_id, name, price, billing_cycle, trial_days, meta, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		RETURNING id, created_at`
	err := r.db.QueryRow(query, plan.ProductID, plan.Name, plan.Price, plan.BillingCycle, plan.TrialDays, plan.Meta).
		Scan(&plan.ID, &plan.CreatedAt)
	if err != nil {
		return fmt.Errorf("create plan: %w", err)
	}
	return nil
}

func (r *PlanRepository) Update(plan *models.Plan) error {
	if plan.Meta == nil {
		plan.Meta = make(models.JSONMap)
	}
	query := `UPDATE plans SET name = $1, price = $2, billing_cycle = $3, trial_days = $4, meta = $5
		WHERE id = $6`
	_, err := r.db.Exec(query, plan.Name, plan.Price, plan.BillingCycle, plan.TrialDays, plan.Meta, plan.ID)
	if err != nil {
		return fmt.Errorf("update plan: %w", err)
	}
	return nil
}

func (r *PlanRepository) DeleteByIDAndBusinessID(id int, businessID string) error {
	query := `DELETE FROM plans pl
		USING products p, services s
		WHERE pl.product_id = p.id AND p.service_id = s.id AND pl.id = $1 AND s.business_id = $2`
	_, err := r.db.Exec(query, id, businessID)
	if err != nil {
		return fmt.Errorf("delete plan: %w", err)
	}
	return nil
}
