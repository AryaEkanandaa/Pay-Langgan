package platformbilling

import (
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"pay-langgan/internal/middlewares"
	service "pay-langgan/internal/services/platformbilling"
	"pay-langgan/internal/utils"
)

type Handler struct{ Service *service.Service }

func (h *Handler) Plan(c echo.Context) error {
	return utils.Success(c, "PayLanggan plan", service.CurrentPlan())
}

func (h *Handler) Status(c echo.Context) error {
	item, err := h.Service.Status(c.Request().Context(), middlewares.GetBusinessID(c))
	if err != nil {
		return handleError(c, err)
	}
	return utils.Success(c, "Platform subscription retrieved", item)
}

func (h *Handler) Invoices(c echo.Context) error {
	items, err := h.Service.Invoices(c.Request().Context(), middlewares.GetBusinessID(c))
	if err != nil {
		return handleError(c, err)
	}
	return utils.Success(c, "Platform invoices retrieved", items)
}

func (h *Handler) Start(c echo.Context) error {
	var body struct {
		Consent bool `json:"auto_renew_consent"`
	}
	if err := c.Bind(&body); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}
	item, err := h.Service.Start(c.Request().Context(), middlewares.GetBusinessID(c), middlewares.GetUserID(c), body.Consent)
	if err != nil {
		return handleError(c, err)
	}
	c.Response().Header().Set("Cache-Control", "no-store")
	return utils.SuccessCreated(c, "Platform checkout created", item)
}

func (h *Handler) Cancel(c echo.Context) error {
	if err := h.Service.Cancel(c.Request().Context(), middlewares.GetBusinessID(c)); err != nil {
		return handleError(c, err)
	}
	return utils.Success(c, "Platform subscription cancelled", nil)
}

func (h *Handler) Webhook(c echo.Context) error {
	c.Request().Body = http.MaxBytesReader(c.Response(), c.Request().Body, 1<<20)
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return utils.BadRequest(c, "invalid webhook body")
	}
	if err = h.Service.Webhook(c.Request().Context(), c.Request().Header.Get("x-callback-token"), body); err != nil {
		return handleError(c, err)
	}
	return utils.Success(c, "Webhook accepted", nil)
}

func handleError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, utils.ErrBadRequest):
		return utils.BadRequest(c, err.Error())
	case errors.Is(err, utils.ErrUnauthorized):
		return utils.Unauthorized(c, "invalid credentials")
	case errors.Is(err, utils.ErrNotFound):
		return utils.NotFound(c, "platform payment not found")
	case errors.Is(err, utils.ErrConflict):
		return utils.Conflict(c, err.Error())
	default:
		slog.Error("platform billing failed", "error", err)
		return utils.InternalError(c, "platform payment processing failed")
	}
}
