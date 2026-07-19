package subscription

import (
	"errors"
	"strconv"

	"pay-langgan/internal/middlewares"
	"pay-langgan/internal/models"
	"pay-langgan/internal/utils"

	"github.com/labstack/echo/v4"
)

func (h *SubscriptionHandler) AddAddOn(c echo.Context) error {
	businessID := middlewares.GetBusinessID(c)
	userID := middlewares.GetUserID(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequest(c, "invalid subscription id")
	}

	var req models.AddSubAddOnRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}

	detail, err := h.subscriptionService.AddAddOn(id, businessID, userID, req.AddOnID, req.Quantity)
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			return utils.NotFound(c, "subscription or add-on not found")
		}
		if errors.Is(err, utils.ErrBadRequest) {
			return utils.BadRequest(c, "quantity must be at least 1")
		}
		return utils.InternalError(c, "failed to add add-on")
	}

	return utils.Success(c, "Add-on added to subscription successfully", detail)
}

func (h *SubscriptionHandler) RemoveAddOn(c echo.Context) error {
	businessID := middlewares.GetBusinessID(c)
	userID := middlewares.GetUserID(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequest(c, "invalid subscription id")
	}
	addOnID, err := strconv.Atoi(c.Param("add_on_id"))
	if err != nil {
		return utils.BadRequest(c, "invalid add-on id")
	}

	detail, err := h.subscriptionService.RemoveAddOn(id, businessID, userID, addOnID)
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			return utils.NotFound(c, "subscription not found")
		}
		return utils.InternalError(c, "failed to remove add-on")
	}

	return utils.Success(c, "Add-on removed from subscription successfully", detail)
}
