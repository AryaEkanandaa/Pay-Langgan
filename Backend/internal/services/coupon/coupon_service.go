package coupon

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"pay-langgan/internal/models"
	"pay-langgan/internal/repositories/coupon"
	"pay-langgan/internal/utils"
)

type CouponService struct {
	couponRepo *coupon.CouponRepository
}

func NewCouponService(couponRepo *coupon.CouponRepository) *CouponService {
	return &CouponService{couponRepo: couponRepo}
}

func (s *CouponService) GetAll(businessID string, page, limit int, search string) ([]models.Coupon, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	return s.couponRepo.FindAll(businessID, page, limit, search)
}

func (s *CouponService) GetByID(id int, businessID string) (*models.Coupon, error) {
	c, err := s.couponRepo.FindByIDAndBusinessID(id, businessID)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, utils.ErrNotFound
	}
	return c, nil
}

func (s *CouponService) Create(businessID string, req models.CreateCouponRequest) (*models.Coupon, error) {
	req.Code = strings.TrimSpace(req.Code)
	if err := validateCoupon(req.Code, req.DiscountType, req.DiscountValue, req.MaxUsage, req.ExpiresAt); err != nil {
		return nil, err
	}

	existing, err := s.couponRepo.FindByCode(businessID, req.Code)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, utils.ErrConflict
	}

	c := &models.Coupon{
		BusinessID:    businessID,
		Code:          req.Code,
		DiscountType:  req.DiscountType,
		DiscountValue: req.DiscountValue,
		MaxUsage:      req.MaxUsage,
		ExpiresAt:     req.ExpiresAt,
	}

	err = s.couponRepo.Create(c)
	if err != nil {
		if errors.Is(err, coupon.ErrDuplicate) {
			return nil, utils.ErrConflict
		}
		return nil, err
	}
	return c, nil
}

func (s *CouponService) Update(id int, businessID string, req models.UpdateCouponRequest) (*models.Coupon, error) {
	c, err := s.couponRepo.FindByIDAndBusinessID(id, businessID)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, utils.ErrNotFound
	}

	if req.Code != "" {
		req.Code = strings.TrimSpace(req.Code)
		if req.Code != c.Code {
			existing, err := s.couponRepo.FindByCode(businessID, req.Code)
			if err != nil {
				return nil, err
			}
			if existing != nil {
				return nil, utils.ErrConflict
			}
		}
		c.Code = req.Code
	}
	if req.DiscountType != "" {
		if req.DiscountType != "percentage" && req.DiscountType != "fixed" {
			return nil, utils.ErrBadRequest
		}
		c.DiscountType = req.DiscountType
	}
	if req.DiscountValue > 0 {
		c.DiscountValue = req.DiscountValue
	}
	if req.MaxUsage != nil {
		c.MaxUsage = req.MaxUsage
	}
	if req.ExpiresAt != nil {
		c.ExpiresAt = req.ExpiresAt
	}
	if err := validateCoupon(c.Code, c.DiscountType, c.DiscountValue, c.MaxUsage, c.ExpiresAt); err != nil {
		return nil, err
	}
	if len(c.Code) > 50 || (c.MaxUsage != nil && *c.MaxUsage < c.UsedCount) {
		return nil, fmt.Errorf("%w: invalid coupon limits", utils.ErrBadRequest)
	}

	err = s.couponRepo.Update(c)
	if err != nil {
		if errors.Is(err, coupon.ErrDuplicate) {
			return nil, utils.ErrConflict
		}
		return nil, err
	}
	return c, nil
}

func (s *CouponService) Delete(id int, businessID string) error {
	c, err := s.couponRepo.FindByIDAndBusinessID(id, businessID)
	if err != nil {
		return err
	}
	if c == nil {
		return utils.ErrNotFound
	}

	return s.couponRepo.Delete(id, businessID)
}

func validateCoupon(code, discountType string, discountValue float64, maxUsage *int, expiresAt *time.Time) error {
	if code == "" {
		return utils.ErrBadRequest
	}
	if len(code) > 50 {
		return fmt.Errorf("%w: code must be at most 50 characters", utils.ErrBadRequest)
	}
	if discountType != "percentage" && discountType != "fixed" {
		return utils.ErrBadRequest
	}
	if discountValue <= 0 {
		return utils.ErrBadRequest
	}
	if discountType == "percentage" && discountValue > 100 {
		return utils.ErrBadRequest
	}
	if maxUsage != nil && *maxUsage < 1 {
		return utils.ErrBadRequest
	}
	if expiresAt != nil && expiresAt.IsZero() {
		return utils.ErrBadRequest
	}
	return nil
}
