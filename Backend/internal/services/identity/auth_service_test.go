package identity

import (
	"errors"
	"testing"

	"pay-langgan/internal/models"
	"pay-langgan/internal/utils"
)

func TestAuthServiceSignupValidation(t *testing.T) {
	service := NewAuthService(nil, nil)
	tests := []struct {
		name string
		req  models.SignupRequest
	}{
		{name: "missing required fields", req: models.SignupRequest{Email: "user@example.com", Password: "secret"}},
		{name: "invalid email", req: models.SignupRequest{BusinessName: "Demo", Name: "Admin", Email: "not-an-email", Password: "secret"}},
		{name: "short password", req: models.SignupRequest{BusinessName: "Demo", Name: "Admin", Email: "user@example.com", Password: "12345"}},
		{name: "business name too long", req: models.SignupRequest{BusinessName: string(make([]byte, 101)), Name: "Admin", Email: "user@example.com", Password: "secret"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.Signup(tt.req)
			if !errors.Is(err, utils.ErrBadRequest) {
				t.Fatalf("Signup() error = %v, want bad request", err)
			}
		})
	}
}

func TestAuthServiceLoginValidation(t *testing.T) {
	service := NewAuthService(nil, nil)
	tests := []models.LoginRequest{
		{},
		{Email: "invalid", Password: "secret"},
		{Email: "user@example.com"},
	}

	for _, req := range tests {
		_, err := service.Login(req)
		if !errors.Is(err, utils.ErrBadRequest) {
			t.Fatalf("Login(%+v) error = %v, want bad request", req, err)
		}
	}
}
