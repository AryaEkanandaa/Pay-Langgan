package payment

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"pay-langgan/internal/middlewares"
	"pay-langgan/internal/models"
	paymentservice "pay-langgan/internal/services/payment"
	"pay-langgan/internal/utils"
)

const maxWebhookBody = 1 << 20

type CheckoutHandler struct {
	service *paymentservice.CheckoutService
}

func NewCheckoutHandler(service *paymentservice.CheckoutService) *CheckoutHandler {
	return &CheckoutHandler{service: service}
}

func (h *CheckoutHandler) RateLimit(write bool) echo.MiddlewareFunc {
	return middlewares.CheckoutRateLimit(h.service, write)
}

func (h *CheckoutHandler) CreateLink(c echo.Context) error {
	invoiceID, err := strconv.Atoi(c.Param("id"))
	if err != nil || invoiceID < 1 {
		return utils.BadRequest(c, "invalid invoice id")
	}
	link, err := h.service.CreateLink(c.Request().Context(), invoiceID, middlewares.GetBusinessID(c))
	if err != nil {
		return checkoutError(c, err)
	}
	return utils.SuccessCreated(c, "Checkout link created", link)
}

func (h *CheckoutHandler) Get(c echo.Context) error {
	checkout, err := h.service.Get(c.Request().Context(), c.Param("token"))
	if err != nil {
		return checkoutError(c, err)
	}
	return utils.Success(c, "Checkout retrieved", checkout)
}

func (h *CheckoutHandler) Start(c echo.Context) error {
	c.Response().Header().Set("Cache-Control", "no-store")
	var request models.StartCheckoutRequest
	if err := c.Bind(&request); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}
	checkout, err := h.service.Start(c.Request().Context(), c.Param("token"), request)
	if err != nil {
		return checkoutError(c, err)
	}
	return utils.Success(c, "Xendit checkout started", checkout)
}

func (h *CheckoutHandler) XenditWebhook(c echo.Context) error {
	request := c.Request()
	request.Body = http.MaxBytesReader(c.Response(), request.Body, maxWebhookBody)
	payload, err := io.ReadAll(request.Body)
	if err != nil {
		return utils.BadRequest(c, "invalid webhook body")
	}
	duplicate, err := h.service.ProcessWebhook(request.Context(), request.Header.Get("x-callback-token"), payload)
	if err != nil {
		return checkoutError(c, err)
	}
	return utils.Success(c, "Webhook accepted", map[string]bool{"duplicate": duplicate})
}

func (h *CheckoutHandler) BusinessGatewayWebhook(c echo.Context) error {
	request := c.Request()
	request.Body = http.MaxBytesReader(c.Response(), request.Body, maxWebhookBody)
	payload, err := io.ReadAll(request.Body)
	if err != nil {
		return utils.BadRequest(c, "invalid webhook body")
	}
	duplicate, err := h.service.ProcessBusinessGatewayWebhook(request.Context(), c.Param("webhookID"), request.Header.Get("x-callback-token"), payload)
	if err != nil {
		return checkoutError(c, err)
	}
	return utils.Success(c, "Webhook accepted", map[string]bool{"duplicate": duplicate})
}

func checkoutError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, utils.ErrNotFound):
		return utils.NotFound(c, "checkout or invoice not found")
	case errors.Is(err, utils.ErrUnauthorized):
		return utils.Unauthorized(c, "invalid webhook token")
	case errors.Is(err, utils.ErrConflict):
		return utils.Conflict(c, err.Error())
	case errors.Is(err, utils.ErrBadRequest):
		return utils.BadRequest(c, err.Error())
	default:
		slog.Error("payment request failed", "error", err)
		return utils.InternalError(c, "payment processing failed")
	}
}
