package customer

import (
	"errors"
	"testing"

	"pay-langgan/internal/models"
	"pay-langgan/internal/utils"
)

func TestCustomerServiceCreateValidation(t *testing.T) {
	service := NewCustomerService(nil)
	tests := []models.CreateCustomerRequest{
		{},
		{Name: "   "},
		{Name: string(make([]byte, 101))},
	}

	for _, req := range tests {
		_, err := service.Create("biz-1", req)
		if !errors.Is(err, utils.ErrBadRequest) {
			t.Fatalf("Create(%+v) error = %v, want bad request", req, err)
		}
	}
}
