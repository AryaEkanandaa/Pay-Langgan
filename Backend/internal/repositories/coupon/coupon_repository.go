package coupon

import (
	"database/sql"
	"fmt"

	"pay-langgan/internal/database"
	"pay-langgan/internal/models"

	"github.com/jmoiron/sqlx"
)

type CouponRepository struct {
	db *database.DB
}

func NewCouponRepository(db *database.DB) *CouponRepository {
	return &CouponRepository{db: db}
}

func (r *CouponRepository) FindAll(businessID string, page, limit int, search string) ([]models.Coupon, int, error) {
	countQuery := `SELECT COUNT(*) FROM coupons WHERE business_id = $1`
	args := []interface{}{businessID}
	argIdx := 2

	if search != "" {
		countQuery += fmt.Sprintf(" AND code ILIKE '%%%%' || $%d || '%%%%'", argIdx)
		args = append(args, search)
		argIdx++
	}

	var total int
	err := r.db.Get(&total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("count coupons: %w", err)
	}

	offset := (page - 1) * limit
	dataQuery := `SELECT id, business_id, code, discount_type, discount_value, max_usage, used_count, expires_at, created_at
		FROM coupons WHERE business_id = $1`
	dataArgs := []interface{}{businessID}
	dataIdx := 2

	if search != "" {
		dataQuery += fmt.Sprintf(" AND code ILIKE '%%%%' || $%d || '%%%%'", dataIdx)
		dataArgs = append(dataArgs, search)
		dataIdx++
	}

	dataQuery += fmt.Sprintf(" ORDER BY id DESC LIMIT $%d OFFSET $%d", dataIdx, dataIdx+1)
	dataArgs = append(dataArgs, limit, offset)

	var coupons []models.Coupon
	err = r.db.Select(&coupons, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("find coupons: %w", err)
	}

	return coupons, total, nil
}

func (r *CouponRepository) FindByIDAndBusinessID(id int, businessID string) (*models.Coupon, error) {
	query := `SELECT id, business_id, code, discount_type, discount_value, max_usage, used_count, expires_at, created_at
		FROM coupons WHERE id = $1 AND business_id = $2`
	var coupon models.Coupon
	err := r.db.Get(&coupon, query, id, businessID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find coupon by id: %w", err)
	}
	return &coupon, nil
}

func (r *CouponRepository) FindByCode(businessID, code string) (*models.Coupon, error) {
	query := `SELECT id, business_id, code, discount_type, discount_value, max_usage, used_count, expires_at, created_at
		FROM coupons WHERE business_id = $1 AND code = $2`
	var coupon models.Coupon
	err := r.db.Get(&coupon, query, businessID, code)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find coupon by code: %w", err)
	}
	return &coupon, nil
}

func (r *CouponRepository) Create(coupon *models.Coupon) error {
	query := `INSERT INTO coupons (business_id, code, discount_type, discount_value, max_usage, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		RETURNING id, created_at`
	err := r.db.QueryRow(query, coupon.BusinessID, coupon.Code, coupon.DiscountType, coupon.DiscountValue, coupon.MaxUsage, coupon.ExpiresAt).
		Scan(&coupon.ID, &coupon.CreatedAt)
	if err != nil {
		if isPGUniqueViolation(err) {
			return fmt.Errorf("%w: coupon code already exists", ErrDuplicate)
		}
		return fmt.Errorf("create coupon: %w", err)
	}
	return nil
}

func (r *CouponRepository) Update(coupon *models.Coupon) error {
	query := `UPDATE coupons SET code = $1, discount_type = $2, discount_value = $3, max_usage = $4, expires_at = $5
		WHERE id = $6 AND business_id = $7`
	_, err := r.db.Exec(query, coupon.Code, coupon.DiscountType, coupon.DiscountValue, coupon.MaxUsage, coupon.ExpiresAt, coupon.ID, coupon.BusinessID)
	if err != nil {
		if isPGUniqueViolation(err) {
			return fmt.Errorf("%w: coupon code already exists", ErrDuplicate)
		}
		return fmt.Errorf("update coupon: %w", err)
	}
	return nil
}

func (r *CouponRepository) UpdateTx(tx *sqlx.Tx, coupon *models.Coupon) error {
	query := `UPDATE coupons SET code = $1, discount_type = $2, discount_value = $3, max_usage = $4, expires_at = $5
		WHERE id = $6 AND business_id = $7`
	_, err := tx.Exec(query, coupon.Code, coupon.DiscountType, coupon.DiscountValue, coupon.MaxUsage, coupon.ExpiresAt, coupon.ID, coupon.BusinessID)
	if err != nil {
		if isPGUniqueViolation(err) {
			return fmt.Errorf("%w: coupon code already exists", ErrDuplicate)
		}
		return fmt.Errorf("update coupon: %w", err)
	}
	return nil
}

func (r *CouponRepository) IncrementUsageTx(tx *sqlx.Tx, id int, businessID string) error {
	query := `UPDATE coupons
		SET used_count = used_count + 1
		WHERE id = $1 AND business_id = $2
		  AND (max_usage IS NULL OR used_count < max_usage)`
	result, err := tx.Exec(query, id, businessID)
	if err != nil {
		return fmt.Errorf("increment coupon usage: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check coupon usage update: %w", err)
	}
	if rows == 0 {
		return ErrUsageLimit
	}
	return nil
}

func (r *CouponRepository) DecrementUsageTx(tx *sqlx.Tx, id int, businessID string) error {
	query := `UPDATE coupons
		SET used_count = GREATEST(used_count - 1, 0)
		WHERE id = $1 AND business_id = $2`
	if _, err := tx.Exec(query, id, businessID); err != nil {
		return fmt.Errorf("decrement coupon usage: %w", err)
	}
	return nil
}

func (r *CouponRepository) Delete(id int, businessID string) error {
	query := `DELETE FROM coupons WHERE id = $1 AND business_id = $2`
	_, err := r.db.Exec(query, id, businessID)
	if err != nil {
		return fmt.Errorf("delete coupon: %w", err)
	}
	return nil
}
