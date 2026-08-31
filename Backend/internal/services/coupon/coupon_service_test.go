package coupon

import (
	"errors"
	"testing"
	"time"

	"pay-langgan/internal/models"
	"pay-langgan/internal/utils"
)

func TestValidateCoupon(t *testing.T) {
	maxUsage := 10
	tests := []struct {
		name          string
		code          string
		discountType  string
		discountValue float64
		maxUsage      *int
		wantBad       bool
	}{
		{name: "valid percentage", code: "DISC10", discountType: "percentage", discountValue: 10, maxUsage: &maxUsage},
		{name: "missing code", discountType: "fixed", discountValue: 100, wantBad: true},
		{name: "invalid type", code: "DISC", discountType: "bogus", discountValue: 10, wantBad: true},
		{name: "percentage above one hundred", code: "DISC", discountType: "percentage", discountValue: 100.01, wantBad: true},
		{name: "non-positive value", code: "DISC", discountType: "fixed", discountValue: 0, wantBad: true},
		{name: "non-positive usage limit", code: "DISC", discountType: "fixed", discountValue: 10, maxUsage: func() *int { v := 0; return &v }(), wantBad: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCoupon(tt.code, tt.discountType, tt.discountValue, tt.maxUsage, nil)
			if tt.wantBad && !errors.Is(err, utils.ErrBadRequest) {
				t.Fatalf("validateCoupon() error = %v, want bad request", err)
			}
			if !tt.wantBad && err != nil {
				t.Fatalf("validateCoupon() unexpected error = %v", err)
			}
		})
	}

	if err := validateCoupon("DISC", "fixed", 10, nil, &time.Time{}); !errors.Is(err, utils.ErrBadRequest) {
		t.Fatalf("zero expiration time should be rejected, got %v", err)
	}
}

func TestCouponServiceCreateValidation(t *testing.T) {
	service := NewCouponService(nil)
	_, err := service.Create("biz-1", models.CreateCouponRequest{
		Code:          "DISC",
		DiscountType:  "percentage",
		DiscountValue: 101,
	})
	if !errors.Is(err, utils.ErrBadRequest) {
		t.Fatalf("Create() error = %v, want bad request", err)
	}
}
