package services

import (
	"fmt"
	"math"
	"time"

	"pay-langgan/internal/models"
	"pay-langgan/internal/repositories"
	"pay-langgan/internal/utils"
)

type SubscriptionPricingService struct {
	planRepo   *repositories.PlanRepository
	addOnRepo  *repositories.AddOnRepository
	couponRepo *repositories.CouponRepository
}

func NewSubscriptionPricingService(planRepo *repositories.PlanRepository, addOnRepo *repositories.AddOnRepository, couponRepo *repositories.CouponRepository) *SubscriptionPricingService {
	return &SubscriptionPricingService{
		planRepo:   planRepo,
		addOnRepo:  addOnRepo,
		couponRepo: couponRepo,
	}
}

func (s *SubscriptionPricingService) CalculatePreview(businessID string, req models.SubscriptionPreviewRequest) (*models.SubscriptionPreviewResponse, error) {
	plan, err := s.planRepo.FindByIDAndBusinessID(req.PlanID, businessID)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		return nil, utils.ErrNotFound
	}

	planAmount := plan.Price
	var items []models.SubscriptionPreviewItem

	items = append(items, models.SubscriptionPreviewItem{
		Type:      "plan",
		Name:      plan.Name,
		Quantity:  1,
		UnitPrice: planAmount,
		Subtotal:  planAmount,
	})

	var addOnAmount float64
	for _, a := range req.AddOns {
		if a.Quantity < 1 {
			continue
		}
		addOn, err := s.addOnRepo.FindByIDAndBusinessID(a.AddOnID, businessID)
		if err != nil {
			return nil, err
		}
		if addOn == nil {
			return nil, fmt.Errorf("add-on %d not found", a.AddOnID)
		}
		subtotal := addOn.Price * float64(a.Quantity)
		addOnAmount += subtotal
		items = append(items, models.SubscriptionPreviewItem{
			Type:      "add_on",
			Name:      addOn.Name,
			Quantity:  a.Quantity,
			UnitPrice: addOn.Price,
			Subtotal:  subtotal,
		})
	}

	subtotalAmount := planAmount + addOnAmount

	var discountAmount float64
	if req.CouponCode != "" {
		coupon, err := s.couponRepo.FindByCode(req.CouponCode)
		if err != nil {
			return nil, err
		}
		if coupon == nil {
			return nil, fmt.Errorf("coupon not found")
		}
		if coupon.ExpiresAt != nil && coupon.ExpiresAt.Before(time.Now()) {
			return nil, fmt.Errorf("coupon has expired")
		}
		if coupon.MaxUsage != nil && coupon.UsedCount >= *coupon.MaxUsage {
			return nil, fmt.Errorf("coupon usage limit exceeded")
		}

		switch coupon.DiscountType {
		case "percentage":
			discountAmount = subtotalAmount * (coupon.DiscountValue / 100)
		case "fixed":
			discountAmount = coupon.DiscountValue
		}
		if discountAmount > subtotalAmount {
			discountAmount = subtotalAmount
		}
		discountAmount = math.Floor(discountAmount*100) / 100

		items = append(items, models.SubscriptionPreviewItem{
			Type:      "discount",
			Name:      fmt.Sprintf("Coupon %s (%s)", req.CouponCode, coupon.DiscountType),
			Quantity:  1,
			UnitPrice: -discountAmount,
			Subtotal:  -discountAmount,
		})
	}

	totalAmount := subtotalAmount - discountAmount
	totalAmount = math.Floor(totalAmount*100) / 100

	return &models.SubscriptionPreviewResponse{
		PlanAmount:     planAmount,
		AddOnAmount:    addOnAmount,
		DiscountAmount: discountAmount,
		TotalAmount:    totalAmount,
		Items:          items,
	}, nil
}
