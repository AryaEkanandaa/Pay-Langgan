package catalog

import (
	"errors"
	"testing"

	"pay-langgan/internal/models"
	"pay-langgan/internal/utils"
)

func TestCatalogCreateValidation(t *testing.T) {
	tests := []struct {
		name string
		call func() error
	}{
		{
			name: "service requires name",
			call: func() error {
				_, err := NewServiceService(nil).Create("biz-1", models.CreateServiceRequest{})
				return err
			},
		},
		{
			name: "product rejects invalid status",
			call: func() error {
				_, err := NewProductService(nil, nil).Create("biz-1", models.CreateProductRequest{
					ServiceID: 1, Name: "Product", Status: "deleted",
				})
				return err
			},
		},
		{
			name: "plan rejects negative trial",
			call: func() error {
				_, err := NewPlanService(nil, nil).Create("biz-1", models.CreatePlanRequest{
					ProductID: 1, Name: "Basic", Price: 10000, BillingCycle: "monthly", TrialDays: -1,
				})
				return err
			},
		},
		{
			name: "add-on rejects invalid billing cycle",
			call: func() error {
				_, err := NewAddOnService(nil, nil).Create("biz-1", models.CreateAddOnRequest{
					ProductID: 1, Name: "Extra", Price: 1000, BillingCycle: "weekly",
				})
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.call(); !errors.Is(err, utils.ErrBadRequest) {
				t.Fatalf("validation error = %v, want bad request", err)
			}
		})
	}
}
