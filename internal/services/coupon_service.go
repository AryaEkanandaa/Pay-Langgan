package services

import (
	"errors"
	"fmt"

	"pay-langgan/internal/models"
	"pay-langgan/internal/repositories"
	"pay-langgan/internal/utils"
)

type CouponService struct {
	couponRepo *repositories.CouponRepository
}

func NewCouponService(couponRepo *repositories.CouponRepository) *CouponService {
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
	coupon, err := s.couponRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if coupon == nil {
		return nil, utils.ErrNotFound
	}
	return coupon, nil
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

	coupon := &models.Coupon{
		Code:          req.Code,
		DiscountType:  req.DiscountType,
		DiscountValue: req.DiscountValue,
		MaxUsage:      req.MaxUsage,
		ExpiresAt:     req.ExpiresAt,
	}

	err = s.couponRepo.Create(coupon)
	if err != nil {
		if errors.Is(err, repositories.ErrDuplicate) {
			return nil, utils.ErrConflict
		}
		return nil, err
	}
	return coupon, nil
}

func (s *CouponService) Update(id int, req models.UpdateCouponRequest) (*models.Coupon, error) {
	coupon, err := s.couponRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if coupon == nil {
		return nil, utils.ErrNotFound
	}

	if req.Code != "" {
		if req.Code != coupon.Code {
			existing, err := s.couponRepo.FindByCode(req.Code)
			if err != nil {
				return nil, err
			}
			if existing != nil {
				return nil, utils.ErrConflict
			}
		}
		coupon.Code = req.Code
	}
	if req.DiscountType != "" {
		if req.DiscountType != "percentage" && req.DiscountType != "fixed" {
			return nil, utils.ErrBadRequest
		}
		coupon.DiscountType = req.DiscountType
	}
	if req.DiscountValue > 0 {
		if req.DiscountType == "percentage" && req.DiscountValue > 100 {
			return nil, fmt.Errorf("discount value cannot exceed 100 for percentage type")
		}
		coupon.DiscountValue = req.DiscountValue
	}
	if req.MaxUsage != nil {
		coupon.MaxUsage = req.MaxUsage
	}
	if req.ExpiresAt != nil {
		coupon.ExpiresAt = req.ExpiresAt
	}

	err = s.couponRepo.Update(coupon)
	if err != nil {
		if errors.Is(err, repositories.ErrDuplicate) {
			return nil, utils.ErrConflict
		}
		return nil, err
	}
	return coupon, nil
}

func (s *CouponService) Delete(id int) error {
	coupon, err := s.couponRepo.FindByID(id)
	if err != nil {
		return err
	}
	if coupon == nil {
		return utils.ErrNotFound
	}

	return s.couponRepo.Delete(id)
}
