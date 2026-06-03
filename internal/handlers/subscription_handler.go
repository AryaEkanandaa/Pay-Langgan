package handlers

import (
	"errors"
	"log"
	"strconv"

	"pay-langgan/internal/middlewares"
	"pay-langgan/internal/models"
	"pay-langgan/internal/services"
	"pay-langgan/internal/utils"

	"github.com/labstack/echo/v4"
)

type SubscriptionHandler struct {
	subscriptionService      *services.SubscriptionService
	subscriptionPricingService *services.SubscriptionPricingService
}

func NewSubscriptionHandler(subscriptionService *services.SubscriptionService, pricingService *services.SubscriptionPricingService) *SubscriptionHandler {
	return &SubscriptionHandler{
		subscriptionService:      subscriptionService,
		subscriptionPricingService: pricingService,
	}
}

func (h *SubscriptionHandler) List(c echo.Context) error {
	businessID := middlewares.GetBusinessID(c)
	page, limit, _ := parsePagination(c)
	status := c.QueryParam("status")
	search := c.QueryParam("search")

	if status != "" && status != "trial" && status != "active" && status != "paused" && status != "cancelled" {
		return utils.BadRequest(c, "invalid status filter")
	}

	subscriptions, total, err := h.subscriptionService.List(businessID, page, limit, status, search)
	if err != nil {
		return utils.InternalError(c, "failed to retrieve subscriptions")
	}

	return utils.SuccessList(c, "Subscriptions retrieved successfully", subscriptions, &models.Pagination{
		Page: page, Limit: limit, Total: total,
	})
}

func (h *SubscriptionHandler) GetByID(c echo.Context) error {
	businessID := middlewares.GetBusinessID(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequest(c, "invalid subscription id")
	}

	detail, err := h.subscriptionService.GetDetail(id, businessID)
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			return utils.NotFound(c, "subscription not found")
		}
		return utils.InternalError(c, "failed to retrieve subscription")
	}

	return utils.Success(c, "Subscription retrieved successfully", detail)
}

func (h *SubscriptionHandler) Create(c echo.Context) error {
	businessID := middlewares.GetBusinessID(c)
	userID := middlewares.GetUserID(c)

	var req models.CreateSubscriptionRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}

	detail, err := h.subscriptionService.Create(businessID, userID, req)
	if err != nil {
		if errors.Is(err, utils.ErrBadRequest) {
			return utils.BadRequest(c, "customer_id and plan_id are required")
		}
		if errors.Is(err, utils.ErrNotFound) {
			return utils.NotFound(c, "customer or plan not found")
		}
		log.Printf("ERROR create subscription: %v", err)
		return utils.InternalError(c, "failed to create subscription")
	}

	return utils.SuccessCreated(c, "Subscription created successfully", detail)
}

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

func (h *SubscriptionHandler) Preview(c echo.Context) error {
	businessID := middlewares.GetBusinessID(c)

	var req models.SubscriptionPreviewRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}

	result, err := h.subscriptionPricingService.CalculatePreview(businessID, req)
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			return utils.NotFound(c, "plan not found")
		}
		return utils.BadRequest(c, err.Error())
	}

	return utils.Success(c, "Subscription preview calculated successfully", result)
}
