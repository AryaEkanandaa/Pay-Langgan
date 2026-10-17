package payment

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

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
	c.Response().Header().Set("Cache-Control", "no-store")
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

func (h *MerchantIntegrationHandler) CreateEmbedSite(c echo.Context) error {
	var body paymentservice.CreateEmbedSiteRequest
	c.Request().Body = http.MaxBytesReader(c.Response(), c.Request().Body, 8<<10)
	if err := c.Bind(&body); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}
	site, err := h.service.CreateEmbedSite(c.Request().Context(), middlewares.GetBusinessID(c), body)
	if err != nil {
		return merchantError(c, err)
	}
	c.Response().Header().Set("Cache-Control", "no-store")
	return utils.SuccessCreated(c, "Embed website created", site)
}

func (h *MerchantIntegrationHandler) ListEmbedSites(c echo.Context) error {
	sites, err := h.service.ListEmbedSites(c.Request().Context(), middlewares.GetBusinessID(c))
	if err != nil {
		return merchantError(c, err)
	}
	return utils.Success(c, "Embed websites retrieved", sites)
}

func (h *MerchantIntegrationHandler) RevokeEmbedSite(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		return utils.BadRequest(c, "invalid website id")
	}
	revoked, err := h.service.RevokeEmbedSite(c.Request().Context(), middlewares.GetBusinessID(c), id)
	if err != nil {
		return merchantError(c, err)
	}
	if !revoked {
		return utils.NotFound(c, "website not found")
	}
	return utils.Success(c, "Embed website revoked", map[string]bool{"revoked": true})
}

func (h *MerchantIntegrationHandler) CreateEmbedOrder(c echo.Context) error {
	c.Response().Header().Set("Cache-Control", "no-store")
	ip, _, err := net.SplitHostPort(c.Request().RemoteAddr)
	if err != nil {
		ip = c.Request().RemoteAddr
	}
	if net.ParseIP(ip) == nil {
		ip = "unknown-peer"
	}
	ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
	defer cancel()
	for _, bucket := range []struct {
		key   string
		limit int
	}{
		{"ip:" + ip, 20}, {"site:" + c.Param("site"), 120},
	} {
		allowed, err := h.service.AllowEmbedRequest(ctx, bucket.key, bucket.limit)
		if err != nil {
			return utils.Error(c, http.StatusServiceUnavailable, "checkout temporarily unavailable")
		}
		if !allowed {
			c.Response().Header().Set("Retry-After", "60")
			return utils.Error(c, http.StatusTooManyRequests, "too many checkout requests; try again shortly")
		}
	}
	var body paymentservice.EmbedOrderRequest
	c.Request().Body = http.MaxBytesReader(c.Response(), c.Request().Body, 8<<10)
	if err := c.Bind(&body); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}
	order, err := h.service.CreateEmbedOrder(c.Request().Context(), c.Param("site"), c.Request().Header.Get("Origin"), body)
	if err != nil {
		return merchantError(c, err)
	}
	return utils.SuccessCreated(c, "Order created", order)
}

func (h *MerchantIntegrationHandler) PreviewEmbedPlan(c echo.Context) error {
	c.Response().Header().Set("Cache-Control", "no-store")
	var options []paymentservice.CheckoutAddOn
	raw := c.QueryParam("add_ons")
	if len(raw) > 4096 {
		return utils.BadRequest(c, "Pilihan add-on tidak valid.")
	}
	if raw != "" {
		if err := json.Unmarshal([]byte(raw), &options); err != nil {
			return utils.BadRequest(c, "Pilihan add-on tidak valid.")
		}
	}
	preview, err := h.service.QuoteEmbedPlan(c.Request().Context(), c.Param("site"), c.Param("plan"), options, c.QueryParam("coupon_code"))
	if err != nil {
		return merchantError(c, err)
	}
	return utils.Success(c, "Package preview", preview)
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
