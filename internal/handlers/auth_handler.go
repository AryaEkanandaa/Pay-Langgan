package handlers

import (
	"errors"
	"net/http"

	"pay-langgan/internal/middlewares"
	"pay-langgan/internal/models"
	"pay-langgan/internal/services"
	"pay-langgan/internal/utils"

	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Signup(c echo.Context) error {
	var req models.SignupRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}

	resp, err := h.authService.Signup(req)
	if err != nil {
		switch {
		case errors.Is(err, utils.ErrBadRequest):
			return utils.BadRequest(c, err.Error())
		case errors.Is(err, utils.ErrConflict):
			return utils.Conflict(c, "email already registered")
		default:
			return utils.InternalError(c, "signup failed")
		}
	}

	return c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "Signup success",
		Data:    resp,
	})
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req models.LoginRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}

	resp, err := h.authService.Login(req)
	if err != nil {
		switch {
		case errors.Is(err, utils.ErrBadRequest):
			return utils.BadRequest(c, "email and password are required")
		case errors.Is(err, utils.ErrUnauthorized):
			return utils.Unauthorized(c, "invalid email or password")
		default:
			return utils.InternalError(c, "login failed")
		}
	}

	return utils.Success(c, "Login success", resp)
}

func (h *AuthHandler) Me(c echo.Context) error {
	userID := middlewares.GetUserID(c)
	businessID := middlewares.GetBusinessID(c)
	email := middlewares.GetEmail(c)
	role := middlewares.GetRole(c)

	resp, err := h.authService.GetMe(userID, businessID, email, role)
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			return utils.NotFound(c, "business not found")
		}
		return utils.InternalError(c, "failed to get profile")
	}

	return utils.Success(c, "User profile retrieved successfully", resp)
}
