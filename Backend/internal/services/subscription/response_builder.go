package subscription

import (
	"database/sql"
	"errors"
	"time"

	"pay-langgan/internal/models"
)

func (s *SubscriptionService) buildDetailResponse(sub *models.Subscription) (*models.SubscriptionDetailResponse, error) {
	resp := &models.SubscriptionDetailResponse{
		ID:         sub.ID,
		CustomerID: sub.CustomerID,
		PlanID:     sub.PlanID,
		Status:     sub.Status,
		StartDate:  sub.StartDate.Format(time.RFC3339),
		Meta:       sub.Meta,
		CreatedAt:  sub.CreatedAt.Format(time.RFC3339),
	}

	if sub.NextBillingDate != nil {
		v := sub.NextBillingDate.Format(time.RFC3339)
		resp.NextBillingDate = &v
	}
	if sub.EndDate != nil {
		v := sub.EndDate.Format(time.RFC3339)
		resp.EndDate = &v
	}
	if sub.TrialEndsAt != nil {
		v := sub.TrialEndsAt.Format(time.RFC3339)
		resp.TrialEndsAt = &v
	}

	cust, err := s.customerRepo.FindByID(sub.CustomerID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if cust != nil {
		resp.Customer = &models.SubscriptionCustomerInfo{
			ID:    cust.ID,
			Name:  cust.Name,
			Email: cust.Email,
		}
	}

	plan, err := s.planRepo.FindByID(sub.PlanID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if plan != nil {
		resp.Plan = &models.SubscriptionPlanInfo{
			ID:           plan.ID,
			Name:         plan.Name,
			Price:        plan.Price,
			BillingCycle: plan.BillingCycle,
			TrialDays:    plan.TrialDays,
		}

		product, err := s.productRepo.FindByID(plan.ProductID)
		if err == nil && product != nil {
			resp.Product = &models.SubscriptionProductInfo{
				ID:   product.ID,
				Name: product.Name,
			}

			service, _ := s.serviceRepo.FindByID(product.ServiceID)
			if service != nil {
				resp.Service = &models.SubscriptionServiceInfo{
					ID:   service.ID,
					Name: service.Name,
				}
			}
		}
	}

	addOnItems, err := s.subAddOnRepo.FindBySubscriptionID(sub.ID)
	if err != nil {
		return nil, err
	}
	for _, a := range addOnItems {
		addOnResp := models.SubscriptionAddOnResponse{
			ID:       a.ID,
			AddOnID:  a.AddOnID,
			Quantity: a.Quantity,
		}
		addOn, _ := s.addOnRepo.FindByID(a.AddOnID)
		if addOn != nil {
			addOnResp.Name = addOn.Name
			addOnResp.Price = addOn.Price
		}
		resp.AddOns = append(resp.AddOns, addOnResp)
	}

	couponItems, err := s.subCpnRepo.FindBySubscriptionID(sub.ID)
	if err != nil {
		return nil, err
	}
	for _, sc := range couponItems {
		coupResp := models.SubscriptionCouponResponse{
			ID:       sc.ID,
			CouponID: sc.CouponID,
		}
		coup, _ := s.couponRepo.FindByID(sc.CouponID)
		if coup != nil {
			coupResp.Code = coup.Code
			coupResp.DiscountType = coup.DiscountType
			coupResp.DiscountValue = coup.DiscountValue
		}
		resp.Coupons = append(resp.Coupons, coupResp)
	}

	return resp, nil
}
