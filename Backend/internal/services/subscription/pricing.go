package subscription

import (
	"fmt"
	"math"
	"strings"
	"time"

	"pay-langgan/internal/models"
	"pay-langgan/internal/repositories/catalog"
	"pay-langgan/internal/repositories/coupon"
	"pay-langgan/internal/utils"
)

type SubscriptionPricingService struct {
	planRepo   *catalog.PlanRepository
	addOnRepo  *catalog.AddOnRepository
	couponRepo *coupon.CouponRepository
}

func NewSubscriptionPricingService(planRepo *catalog.PlanRepository, addOnRepo *catalog.AddOnRepository, couponRepo *coupon.CouponRepository) *SubscriptionPricingService {
	return &SubscriptionPricingService{
		planRepo:   planRepo,
		addOnRepo:  addOnRepo,
		couponRepo: couponRepo,
	}
}

func (s *SubscriptionPricingService) CalculatePreview(businessID string, req models.SubscriptionPreviewRequest) (*models.SubscriptionPreviewResponse, error) {
	req.CouponCode = strings.TrimSpace(req.CouponCode)
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
		if a.AddOnID < 1 || a.Quantity < 1 {
			return nil, utils.ErrBadRequest
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
		coup, err := s.couponRepo.FindByCode(businessID, req.CouponCode)
		if err != nil {
			return nil, err
		}
		if coup == nil {
			return nil, fmt.Errorf("coupon not found")
		}
		if coup.ExpiresAt != nil && coup.ExpiresAt.Before(time.Now()) {
			return nil, fmt.Errorf("coupon has expired")
		}
		if coup.MaxUsage != nil && coup.UsedCount >= *coup.MaxUsage {
			return nil, fmt.Errorf("coupon usage limit exceeded")
		}

		switch coup.DiscountType {
		case "percentage":
			discountAmount = subtotalAmount * (coup.DiscountValue / 100)
		case "fixed":
			discountAmount = coup.DiscountValue
		}
		if discountAmount > subtotalAmount {
			discountAmount = subtotalAmount
		}
		discountAmount = math.Floor(discountAmount*100) / 100

		items = append(items, models.SubscriptionPreviewItem{
			Type:      "discount",
			Name:      fmt.Sprintf("Coupon %s (%s)", req.CouponCode, coup.DiscountType),
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
