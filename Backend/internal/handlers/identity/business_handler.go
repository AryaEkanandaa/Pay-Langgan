package identity

import (
	"errors"
	"net/http"

	"pay-langgan/internal/middlewares"
	"pay-langgan/internal/models"
	"pay-langgan/internal/services/identity"
	"pay-langgan/internal/utils"

	"github.com/labstack/echo/v4"
)

type BusinessHandler struct {
	businessService *identity.BusinessService
}

func NewBusinessHandler(businessService *identity.BusinessService) *BusinessHandler {
	return &BusinessHandler{businessService: businessService}
}

func (h *BusinessHandler) GetMyBusiness(c echo.Context) error {
	businessID := middlewares.GetBusinessID(c)

	business, err := h.businessService.GetMyBusiness(businessID)
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			return utils.NotFound(c, "business not found")
		}
		return utils.InternalError(c, "failed to retrieve business")
	}

	return utils.Success(c, "Business retrieved successfully", business)
}

func (h *BusinessHandler) UpdateMyBusiness(c echo.Context) error {
	businessID := middlewares.GetBusinessID(c)

	var req models.UpdateBusinessRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}

	err := h.businessService.UpdateMyBusiness(businessID, req)
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			return utils.NotFound(c, "business not found")
		}
		return utils.InternalError(c, "failed to update business")
	}

	return c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Business updated successfully",
	})
}
