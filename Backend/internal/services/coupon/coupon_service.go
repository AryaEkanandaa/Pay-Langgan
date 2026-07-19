package coupon

import (
	"errors"
	"fmt"

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

func (s *CouponService) GetAll(page, limit int, search string) ([]models.Coupon, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	return s.couponRepo.FindAll(page, limit, search)
}

func (s *CouponService) GetByID(id int) (*models.Coupon, error) {
	c, err := s.couponRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, utils.ErrNotFound
	}
	return c, nil
}

func (s *CouponService) Create(req models.CreateCouponRequest) (*models.Coupon, error) {
	if req.Code == "" {
		return nil, utils.ErrBadRequest
	}
	if req.DiscountType != "percentage" && req.DiscountType != "fixed" {
		return nil, utils.ErrBadRequest
	}
	if req.DiscountValue <= 0 {
		return nil, utils.ErrBadRequest
	}
	if req.DiscountType == "percentage" && req.DiscountValue > 100 {
		return nil, fmt.Errorf("discount value cannot exceed 100 for percentage type")
	}

	existing, err := s.couponRepo.FindByCode(req.Code)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, utils.ErrConflict
	}

	c := &models.Coupon{
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

func (s *CouponService) Update(id int, req models.UpdateCouponRequest) (*models.Coupon, error) {
	c, err := s.couponRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, utils.ErrNotFound
	}

	if req.Code != "" {
		if req.Code != c.Code {
			existing, err := s.couponRepo.FindByCode(req.Code)
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
		if req.DiscountType == "percentage" && req.DiscountValue > 100 {
			return nil, fmt.Errorf("discount value cannot exceed 100 for percentage type")
		}
		c.DiscountValue = req.DiscountValue
	}
	if req.MaxUsage != nil {
		c.MaxUsage = req.MaxUsage
	}
	if req.ExpiresAt != nil {
		c.ExpiresAt = req.ExpiresAt
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

func (s *CouponService) Delete(id int) error {
	c, err := s.couponRepo.FindByID(id)
	if err != nil {
		return err
	}
	if c == nil {
		return utils.ErrNotFound
	}

	return s.couponRepo.Delete(id)
}
