package identity

import (
	"errors"

	"pay-langgan/internal/middlewares"
	"pay-langgan/internal/models"
	identitysvc "pay-langgan/internal/services/identity"
	"pay-langgan/internal/utils"

	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	userService *identitysvc.UserService
}

func NewUserHandler(userService *identitysvc.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) List(c echo.Context) error {
	users, err := h.userService.ListBusinessUsers(middlewares.GetBusinessID(c))
	if err != nil {
		return utils.InternalError(c, "failed to retrieve users")
	}
	return utils.Success(c, "Users retrieved successfully", users)
}

func (h *UserHandler) CreateStaff(c echo.Context) error {
	var req models.CreateStaffRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}

	user, err := h.userService.CreateStaff(middlewares.GetBusinessID(c), req)
	if err != nil {
		switch {
		case errors.Is(err, utils.ErrBadRequest):
			return utils.BadRequest(c, err.Error())
		case errors.Is(err, utils.ErrConflict):
			return utils.Conflict(c, "email already registered")
		default:
			return utils.InternalError(c, "failed to create user")
		}
	}
	return utils.SuccessCreated(c, "Staff user created successfully", user)
}
