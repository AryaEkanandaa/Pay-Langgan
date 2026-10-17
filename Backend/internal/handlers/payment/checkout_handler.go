package payment

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

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
	c.Response().Header().Set("Cache-Control", "no-store")
	checkout, err := h.service.Get(c.Request().Context(), c.Param("token"))
	if err != nil {
		return checkoutError(c, err)
	}
	return utils.Success(c, "Checkout retrieved", checkout)
}

func (h *CheckoutHandler) Cancel(c echo.Context) error {
	c.Response().Header().Set("Cache-Control", "no-store")
	checkout, err := h.service.Cancel(c.Request().Context(), c.Param("token"))
	if err != nil {
		return checkoutError(c, err)
	}
	return utils.Success(c, "Checkout cancelled", checkout)
}

func (h *CheckoutHandler) ListWebhookLogs(c echo.Context) error {
	c.Response().Header().Set("Cache-Control", "no-store")
	page := 1
	if value := c.QueryParam("page"); value != "" {
		var err error
		page, err = strconv.Atoi(value)
		if err != nil {
			return utils.BadRequest(c, "invalid page")
		}
	}
	logs, err := h.service.ListWebhookLogs(c.Request().Context(), middlewares.GetBusinessID(c), c.QueryParam("provider"), c.QueryParam("status"), page)
	if err != nil {
		return checkoutError(c, err)
	}
	return utils.Success(c, "Webhook logs retrieved", logs)
}

func (h *CheckoutHandler) NewWebhookCount(c echo.Context) error {
	c.Response().Header().Set("Cache-Control", "no-store")
	count, err := h.service.NewWebhookCount(c.Request().Context(), middlewares.GetBusinessID(c), middlewares.GetUserID(c))
	if err != nil {
		return checkoutError(c, err)
	}
	return utils.Success(c, "New webhook count", count)
}

func (h *CheckoutHandler) MarkWebhookLogsSeen(c echo.Context) error {
	c.Response().Header().Set("Cache-Control", "no-store")
	var request struct {
		SeenThrough time.Time `json:"seen_through"`
	}
	if err := c.Bind(&request); err != nil {
		return utils.BadRequest(c, "invalid seen timestamp")
	}
	count, err := h.service.MarkWebhookLogsSeen(c.Request().Context(), middlewares.GetBusinessID(c), middlewares.GetUserID(c), request.SeenThrough)
	if err != nil {
		return checkoutError(c, err)
	}
	return utils.Success(c, "Webhook logs marked read", count)
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
	return utils.Success(c, "Checkout started", checkout)
}

func (h *CheckoutHandler) ChargeMidtransCard(c echo.Context) error {
	c.Response().Header().Set("Cache-Control", "no-store")
	c.Request().Body = http.MaxBytesReader(c.Response(), c.Request().Body, 4096)
	var request models.MidtransCardCheckoutRequest
	decoder := json.NewDecoder(c.Request().Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return utils.BadRequest(c, "invalid card token request")
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return utils.BadRequest(c, "invalid card token request")
	}
	result, err := h.service.ChargeMidtransCard(c.Request().Context(), c.Param("token"), request)
	if err != nil {
		return checkoutError(c, err)
	}
	return utils.Success(c, "Card payment started", result)
}

func (h *CheckoutHandler) VerifyMidtransCard(c echo.Context) error {
	c.Response().Header().Set("Cache-Control", "no-store")
	checkout, err := h.service.VerifyMidtransCard(c.Request().Context(), c.Param("token"))
	if err != nil {
		return checkoutError(c, err)
	}
	return utils.Success(c, "Payment status verified", checkout)
}

func (h *CheckoutHandler) RegisterMidtransTrialCard(c echo.Context) error {
	c.Response().Header().Set("Cache-Control", "no-store")
	c.Request().Body = http.MaxBytesReader(c.Response(), c.Request().Body, 1024)
	var request paymentservice.MidtransTrialRegistration
	decoder := json.NewDecoder(c.Request().Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return utils.BadRequest(c, "invalid registration receipt")
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return utils.BadRequest(c, "invalid registration receipt")
	}
	result, err := h.service.RegisterMidtransTrialCard(c.Request().Context(), c.Param("token"), request)
	if err != nil {
		return checkoutError(c, err)
	}
	return utils.Success(c, "Card registration recorded", result)
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
