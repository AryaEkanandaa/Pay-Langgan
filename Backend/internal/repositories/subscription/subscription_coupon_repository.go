package subscription

import (
	"fmt"

	"pay-langgan/internal/database"
	"pay-langgan/internal/models"

	"github.com/jmoiron/sqlx"
)

type SubscriptionCouponRepository struct {
	db *database.DB
}

func NewSubscriptionCouponRepository(db *database.DB) *SubscriptionCouponRepository {
	return &SubscriptionCouponRepository{db: db}
}

func (r *SubscriptionCouponRepository) Apply(tx *sqlx.Tx, item *models.SubscriptionCoupon) error {
	query := `INSERT INTO subscription_coupons (subscription_id, coupon_id, applied_at, expired_at, created_at)
		VALUES ($1, $2, NOW(), $3, NOW())
		RETURNING id, created_at`
	err := tx.QueryRow(query, item.SubscriptionID, item.CouponID, item.ExpiredAt).
		Scan(&item.ID, &item.CreatedAt)
	if err != nil {
		return fmt.Errorf("apply subscription coupon: %w", err)
	}
	return nil
}

func (r *SubscriptionCouponRepository) RemoveBySubscriptionIDAndCouponID(tx *sqlx.Tx, subscriptionID, couponID int) error {
	query := `DELETE FROM subscription_coupons WHERE subscription_id = $1 AND coupon_id = $2`
	_, err := tx.Exec(query, subscriptionID, couponID)
	if err != nil {
		return fmt.Errorf("remove subscription coupon: %w", err)
	}
	return nil
}

func (r *SubscriptionCouponRepository) FindBySubscriptionID(subscriptionID int) ([]models.SubscriptionCoupon, error) {
	query := `SELECT id, subscription_id, coupon_id, applied_at, expired_at, created_at
		FROM subscription_coupons WHERE subscription_id = $1`
	var items []models.SubscriptionCoupon
	err := r.db.Select(&items, query, subscriptionID)
	if err != nil {
		return nil, fmt.Errorf("find subscription coupons: %w", err)
	}
	return items, nil
}

func (r *SubscriptionCouponRepository) ExistsBySubscriptionIDAndCouponID(subscriptionID, couponID int) (bool, error) {
	query := `SELECT COUNT(*) FROM subscription_coupons WHERE subscription_id = $1 AND coupon_id = $2`
	var count int
	err := r.db.Get(&count, query, subscriptionID, couponID)
	if err != nil {
		return false, fmt.Errorf("check subscription coupon exists: %w", err)
	}
	return count > 0, nil
}
