package subscription

import (
	"errors"
	"strconv"

	"pay-langgan/internal/middlewares"
	"pay-langgan/internal/models"
	"pay-langgan/internal/utils"

	"github.com/labstack/echo/v4"
)

func (h *SubscriptionHandler) Cancel(c echo.Context) error {
	businessID := middlewares.GetBusinessID(c)
	userID := middlewares.GetUserID(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequest(c, "invalid subscription id")
	}

	var req models.CancelSubscriptionRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}

	detail, err := h.subscriptionService.Cancel(id, businessID, userID, req.Reason)
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			return utils.NotFound(c, "subscription not found")
		}
		return utils.BadRequest(c, err.Error())
	}

	return utils.Success(c, "Subscription cancelled successfully", detail)
}

func (h *SubscriptionHandler) Pause(c echo.Context) error {
	businessID := middlewares.GetBusinessID(c)
	userID := middlewares.GetUserID(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequest(c, "invalid subscription id")
	}

	var req models.PauseSubscriptionRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}

	detail, err := h.subscriptionService.Pause(id, businessID, userID, req.Reason)
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			return utils.NotFound(c, "subscription not found")
		}
		return utils.BadRequest(c, err.Error())
	}

	return utils.Success(c, "Subscription paused successfully", detail)
}

func (h *SubscriptionHandler) Resume(c echo.Context) error {
	businessID := middlewares.GetBusinessID(c)
	userID := middlewares.GetUserID(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequest(c, "invalid subscription id")
	}

	detail, err := h.subscriptionService.Resume(id, businessID, userID)
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			return utils.NotFound(c, "subscription not found")
		}
		return utils.BadRequest(c, err.Error())
	}

	return utils.Success(c, "Subscription resumed successfully", detail)
}
