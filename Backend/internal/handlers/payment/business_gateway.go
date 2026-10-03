package payment

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"pay-langgan/internal/middlewares"
	paymentservice "pay-langgan/internal/services/payment"
	"pay-langgan/internal/utils"
)

type BusinessGatewayHandler struct {
	service *paymentservice.BusinessGatewayService
}

func NewBusinessGatewayHandler(service *paymentservice.BusinessGatewayService) *BusinessGatewayHandler {
	return &BusinessGatewayHandler{service: service}
}

func (h *BusinessGatewayHandler) Get(c echo.Context) error {
	c.Response().Header().Set("Cache-Control", "no-store")
	item, err := h.service.Get(c.Request().Context(), middlewares.GetBusinessID(c))
	if errors.Is(err, utils.ErrNotFound) {
		return utils.Success(c, "No payment gateway configured", nil)
	}
	if err != nil {
		return businessGatewayError(c, err)
	}
	return utils.Success(c, "Payment gateway configuration retrieved", item)
}

func (h *BusinessGatewayHandler) Save(c echo.Context) error {
	c.Response().Header().Set("Cache-Control", "no-store")
	var body paymentservice.BusinessGatewayInput
	c.Request().Body = http.MaxBytesReader(c.Response(), c.Request().Body, 16<<10)
	if err := c.Bind(&body); err != nil {
		return utils.BadRequest(c, "invalid payment gateway configuration")
	}
	item, err := h.service.Save(c.Request().Context(), middlewares.GetBusinessID(c), body)
	if err != nil {
		return businessGatewayError(c, err)
	}
	return utils.Success(c, "Payment gateway configuration saved", item)
}

func (h *BusinessGatewayHandler) Delete(c echo.Context) error {
	if err := h.service.Delete(c.Request().Context(), middlewares.GetBusinessID(c)); err != nil {
		return businessGatewayError(c, err)
	}
	return utils.Success(c, "Payment gateway configuration removed", map[string]bool{"removed": true})
}

func businessGatewayError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, utils.ErrBadRequest):
		return utils.BadRequest(c, err.Error())
	case errors.Is(err, utils.ErrConflict):
		return utils.Conflict(c, err.Error())
	case errors.Is(err, utils.ErrNotFound):
		return utils.NotFound(c, "payment gateway configuration not found")
	default:
		slog.Error("business payment gateway request failed", "error", err)
		return utils.InternalError(c, "payment gateway configuration request failed")
	}
}
