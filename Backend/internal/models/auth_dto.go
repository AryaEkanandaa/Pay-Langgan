package models

type SignupRequest struct {
	BusinessName string `json:"business_name" validate:"required"`
	Name         string `json:"name" validate:"required"`
	Email        string `json:"email" validate:"required,email"`
	Password     string `json:"password" validate:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	Token    string      `json:"token"`
	User     UserDTO     `json:"user"`
	Business BusinessDTO `json:"business"`
}

type LoginResponse struct {
	Token string  `json:"token"`
	User  UserDTO `json:"user"`
}

type MeResponse struct {
	UserID     int    `json:"user_id"`
	BusinessID string `json:"business_id"`
	Email      string `json:"email"`
	Role       string `json:"role"`
}

type UserDTO struct {
	ID         int    `json:"id"`
	BusinessID string `json:"business_id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Role       string `json:"role"`
}

type CreateStaffRequest struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Role     string `json:"role" validate:"required"`
}
