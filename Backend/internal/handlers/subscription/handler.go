package subscription

import (
	"errors"
	"log"
	"strconv"

	"pay-langgan/internal/middlewares"
	"pay-langgan/internal/models"
	"pay-langgan/internal/services/subscription"
	"pay-langgan/internal/utils"

	"github.com/labstack/echo/v4"
)

type SubscriptionHandler struct {
	subscriptionService        *subscription.SubscriptionService
	subscriptionPricingService *subscription.SubscriptionPricingService
}

func NewSubscriptionHandler(subscriptionService *subscription.SubscriptionService, pricingService *subscription.SubscriptionPricingService) *SubscriptionHandler {
	return &SubscriptionHandler{
		subscriptionService:        subscriptionService,
		subscriptionPricingService: pricingService,
	}
}

func (h *SubscriptionHandler) List(c echo.Context) error {
	businessID := middlewares.GetBusinessID(c)
	page, limit, _ := utils.ParsePagination(c)
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
