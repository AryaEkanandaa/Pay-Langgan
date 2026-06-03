package repositories

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

func (r *CouponRepository) FindAll(page, limit int, search string) ([]models.Coupon, int, error) {
	countQuery := `SELECT COUNT(*) FROM coupons`
	args := []interface{}{}
	argIdx := 1

	if search != "" {
		countQuery += fmt.Sprintf(" WHERE code ILIKE '%%%%' || $%d || '%%%%'", argIdx)
		args = append(args, search)
		argIdx++
	}

	var total int
	err := r.db.Get(&total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("count coupons: %w", err)
	}

	offset := (page - 1) * limit
	dataQuery := `SELECT id, code, discount_type, discount_value, max_usage, used_count, expires_at, created_at
		FROM coupons`
	dataArgs := []interface{}{}
	dataIdx := 1

	if search != "" {
		dataQuery += fmt.Sprintf(" WHERE code ILIKE '%%%%' || $%d || '%%%%'", dataIdx)
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

func (r *CouponRepository) FindByID(id int) (*models.Coupon, error) {
	query := `SELECT id, code, discount_type, discount_value, max_usage, used_count, expires_at, created_at
		FROM coupons WHERE id = $1`
	var coupon models.Coupon
	err := r.db.Get(&coupon, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find coupon by id: %w", err)
	}
	return &coupon, nil
}

func (r *CouponRepository) FindByCode(code string) (*models.Coupon, error) {
	query := `SELECT id, code, discount_type, discount_value, max_usage, used_count, expires_at, created_at
		FROM coupons WHERE code = $1`
	var coupon models.Coupon
	err := r.db.Get(&coupon, query, code)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find coupon by code: %w", err)
	}
	return &coupon, nil
}

func (r *CouponRepository) Create(coupon *models.Coupon) error {
	query := `INSERT INTO coupons (code, discount_type, discount_value, max_usage, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING id, created_at`
	err := r.db.QueryRow(query, coupon.Code, coupon.DiscountType, coupon.DiscountValue, coupon.MaxUsage, coupon.ExpiresAt).
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
		WHERE id = $6`
	_, err := r.db.Exec(query, coupon.Code, coupon.DiscountType, coupon.DiscountValue, coupon.MaxUsage, coupon.ExpiresAt, coupon.ID)
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
		WHERE id = $6`
	_, err := tx.Exec(query, coupon.Code, coupon.DiscountType, coupon.DiscountValue, coupon.MaxUsage, coupon.ExpiresAt, coupon.ID)
	if err != nil {
		if isPGUniqueViolation(err) {
			return fmt.Errorf("%w: coupon code already exists", ErrDuplicate)
		}
		return fmt.Errorf("update coupon: %w", err)
	}
	return nil
}

func (r *CouponRepository) Delete(id int) error {
	query := `DELETE FROM coupons WHERE id = $1`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("delete coupon: %w", err)
	}
	return nil
}
