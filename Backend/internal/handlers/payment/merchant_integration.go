package payment

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"pay-langgan/internal/middlewares"
	paymentservice "pay-langgan/internal/services/payment"
	"pay-langgan/internal/utils"
)

type MerchantIntegrationHandler struct {
	service *paymentservice.MerchantIntegrationService
}

func NewMerchantIntegrationHandler(service *paymentservice.MerchantIntegrationService) *MerchantIntegrationHandler {
	return &MerchantIntegrationHandler{service: service}
}

func (h *MerchantIntegrationHandler) CreateKey(c echo.Context) error {
	c.Response().Header().Set("Cache-Control", "no-store")
	var body struct {
		Name string `json:"name"`
	}
	if err := c.Bind(&body); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}
	key, err := h.service.CreateKey(c.Request().Context(), middlewares.GetBusinessID(c), body.Name)
	if err != nil {
		return merchantError(c, err)
	}
	return utils.SuccessCreated(c, "Integration key created; save it now because it will not be shown again", key)
}

func (h *MerchantIntegrationHandler) ListKeys(c echo.Context) error {
	keys, err := h.service.ListKeys(c.Request().Context(), middlewares.GetBusinessID(c))
	if err != nil {
		return merchantError(c, err)
	}
	return utils.Success(c, "Integration keys retrieved", keys)
}

func (h *MerchantIntegrationHandler) RevokeKey(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		return utils.BadRequest(c, "invalid key id")
	}
	revoked, err := h.service.RevokeKey(c.Request().Context(), middlewares.GetBusinessID(c), id)
	if err != nil {
		return merchantError(c, err)
	}
	if !revoked {
		return utils.NotFound(c, "integration key not found")
	}
	return utils.Success(c, "Integration key revoked", map[string]bool{"revoked": true})
}

func (h *MerchantIntegrationHandler) businessFromKey(c echo.Context) (string, error) {
	authorization := strings.TrimSpace(c.Request().Header.Get("Authorization"))
	if !strings.HasPrefix(authorization, "Bearer ") {
		return "", utils.ErrUnauthorized
	}
	return h.service.Authenticate(c.Request().Context(), strings.TrimSpace(strings.TrimPrefix(authorization, "Bearer ")))
}

func (h *MerchantIntegrationHandler) CreateOrder(c echo.Context) error {
	c.Response().Header().Set("Cache-Control", "no-store")
	businessID, err := h.businessFromKey(c)
	if err != nil {
		return merchantError(c, err)
	}
	var request paymentservice.MerchantOrderRequest
	c.Request().Body = http.MaxBytesReader(c.Response(), c.Request().Body, 8<<10)
	if err := c.Bind(&request); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}
	order, err := h.service.CreateOrder(c.Request().Context(), businessID, request)
	if err != nil {
		return merchantError(c, err)
	}
	return utils.SuccessCreated(c, "Order created", order)
}

func (h *MerchantIntegrationHandler) GetOrder(c echo.Context) error {
	c.Response().Header().Set("Cache-Control", "no-store")
	businessID, err := h.businessFromKey(c)
	if err != nil {
		return merchantError(c, err)
	}
	order, err := h.service.GetOrder(c.Request().Context(), businessID, c.Param("reference"))
	if err != nil {
		return merchantError(c, err)
	}
	return utils.Success(c, "Order retrieved", order)
}

func merchantError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, utils.ErrUnauthorized):
		return utils.Unauthorized(c, "invalid integration key")
	case errors.Is(err, utils.ErrForbidden):
		return utils.Forbidden(c, err.Error())
	case errors.Is(err, utils.ErrBadRequest):
		return utils.BadRequest(c, err.Error())
	case errors.Is(err, utils.ErrNotFound):
		return utils.NotFound(c, "order or plan not found")
	case errors.Is(err, utils.ErrConflict):
		return utils.Conflict(c, err.Error())
	default:
		slog.Error("merchant integration request failed", "error", err)
		return utils.InternalError(c, "merchant request failed")
	}
}
