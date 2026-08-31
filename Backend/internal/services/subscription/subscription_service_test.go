package subscription

import (
	"errors"
	"testing"

	"pay-langgan/internal/models"
	"pay-langgan/internal/utils"
)

func TestSubscriptionServiceCreateValidation(t *testing.T) {
	service := NewSubscriptionService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	tests := []struct {
		name string
		req  models.CreateSubscriptionRequest
	}{
		{name: "missing customer", req: models.CreateSubscriptionRequest{PlanID: 1}},
		{name: "missing plan", req: models.CreateSubscriptionRequest{CustomerID: 1}},
		{name: "invalid add-on id", req: models.CreateSubscriptionRequest{
			CustomerID: 1, PlanID: 1, AddOns: []models.CreateSubAddOnItem{{AddOnID: 0, Quantity: 1}},
		}},
		{name: "invalid add-on quantity", req: models.CreateSubscriptionRequest{
			CustomerID: 1, PlanID: 1, AddOns: []models.CreateSubAddOnItem{{AddOnID: 1, Quantity: 0}},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.Create("biz-1", 1, tt.req)
			if !errors.Is(err, utils.ErrBadRequest) {
				t.Fatalf("Create() error = %v, want bad request", err)
			}
		})
	}
}

func TestSubscriptionServiceActionValidation(t *testing.T) {
	service := NewSubscriptionService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	if _, err := service.AddAddOn(1, "biz-1", 1, 0, 1); !errors.Is(err, utils.ErrBadRequest) {
		t.Fatalf("AddAddOn() error = %v, want bad request", err)
	}
	if _, err := service.AddAddOn(1, "biz-1", 1, 1, 0); !errors.Is(err, utils.ErrBadRequest) {
		t.Fatalf("AddAddOn() error = %v, want bad request", err)
	}
	if _, err := service.ApplyCoupon(1, "biz-1", 1, "   "); !errors.Is(err, utils.ErrBadRequest) {
		t.Fatalf("ApplyCoupon() error = %v, want bad request", err)
	}
}
