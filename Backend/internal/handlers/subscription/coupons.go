package subscription

import (
	"errors"
	"strconv"

	"pay-langgan/internal/middlewares"
	"pay-langgan/internal/models"
	"pay-langgan/internal/utils"

	"github.com/labstack/echo/v4"
)

func (h *SubscriptionHandler) ApplyCoupon(c echo.Context) error {
	businessID := middlewares.GetBusinessID(c)
	userID := middlewares.GetUserID(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequest(c, "invalid subscription id")
	}

	var req models.ApplyCouponRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}

	detail, err := h.subscriptionService.ApplyCoupon(id, businessID, userID, req.CouponCode)
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			return utils.NotFound(c, "subscription or coupon not found")
		}
		if errors.Is(err, utils.ErrConflict) {
			return utils.Conflict(c, "coupon already applied to this subscription")
		}
		return utils.BadRequest(c, err.Error())
	}

	return utils.Success(c, "Coupon applied to subscription successfully", detail)
}

func (h *SubscriptionHandler) RemoveCoupon(c echo.Context) error {
	businessID := middlewares.GetBusinessID(c)
	userID := middlewares.GetUserID(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequest(c, "invalid subscription id")
	}
	couponID, err := strconv.Atoi(c.Param("coupon_id"))
	if err != nil {
		return utils.BadRequest(c, "invalid coupon id")
	}

	detail, err := h.subscriptionService.RemoveCoupon(id, businessID, userID, couponID)
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			return utils.NotFound(c, "subscription or coupon not found")
		}
		return utils.InternalError(c, "failed to remove coupon")
	}

	return utils.Success(c, "Coupon removed from subscription successfully", detail)
}
