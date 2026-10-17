package portal

import (
	"errors"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
	"net"
	"net/http"
	"pay-langgan/internal/services/billing"
	"pay-langgan/internal/services/payment"
	service "pay-langgan/internal/services/portal"
	"pay-langgan/internal/utils"
	"strconv"
	"strings"
	"time"
)

type Handler struct {
	Service  *service.Service
	Invoice  *billing.InvoiceService
	Checkout *payment.CheckoutService
}

func (h *Handler) RateLimit() echo.MiddlewareFunc {
	return middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{Store: middleware.NewRateLimiterMemoryStoreWithConfig(middleware.RateLimiterMemoryStoreConfig{Rate: rate.Limit(1.0 / 3), Burst: 20, ExpiresIn: 5 * time.Minute}), IdentifierExtractor: func(c echo.Context) (string, error) {
		host, _, e := net.SplitHostPort(c.Request().RemoteAddr)
		if e != nil {
			return c.Request().RemoteAddr, nil
		}
		return host, nil
	}, DenyHandler: func(c echo.Context, _ string, _ error) error {
		return c.JSON(429, map[string]interface{}{"success": false, "message": "Terlalu banyak permintaan. Tunggu sebentar lalu coba lagi."})
	}})
}
func (h *Handler) Headers(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("Cache-Control", "no-store")
		c.Response().Header().Set("Referrer-Policy", "no-referrer")
		c.Request().Body = http.MaxBytesReader(c.Response(), c.Request().Body, 4096)
		return next(c)
	}
}
func portalError(c echo.Context, e error) error {
	if errors.Is(e, utils.ErrUnauthorized) {
		return utils.Unauthorized(c, "Sesi atau tautan tidak valid, kedaluwarsa, atau sudah digunakan. Minta tautan masuk baru.")
	}
	if errors.Is(e, utils.ErrNotFound) {
		return utils.NotFound(c, "Data portal tidak ditemukan.")
	}
	if errors.Is(e, utils.ErrBadRequest) {
		return utils.BadRequest(c, "Data permintaan tidak valid.")
	}
	if errors.Is(e, utils.ErrConflict) {
		return c.JSON(409, map[string]interface{}{"success": false, "message": "Pembayaran sedang diproses atau invoice tidak lagi dapat dibayar. Perbarui status atau hubungi bisnis."})
	}
	return utils.InternalError(c, "Portal belum dapat diproses. Coba lagi nanti.")
}
func (h *Handler) Business(c echo.Context) error {
	b, e := h.Service.Business(c.Request().Context(), c.Param("site"))
	if e != nil {
		return portalError(c, e)
	}
	return utils.Success(c, "Portal pelanggan", b)
}
func (h *Handler) Request(c echo.Context) error {
	started := time.Now()
	var req struct {
		Email string `json:"email"`
	}
	if e := c.Bind(&req); e != nil {
		return utils.BadRequest(c, "Masukkan alamat email yang valid.")
	}
	e := h.Service.RequestLink(c.Request().Context(), c.Param("site"), req.Email)
	timer := time.NewTimer(max(0, 200*time.Millisecond-time.Since(started)))
	defer timer.Stop()
	select {
	case <-timer.C:
	case <-c.Request().Context().Done():
		return nil
	}
	if e != nil {
		return portalError(c, e)
	}
	return utils.Success(c, "Jika email terdaftar pada bisnis ini, tautan masuk akan dikirim. Periksa inbox dan folder spam.", nil)
}
func (h *Handler) Exchange(c echo.Context) error {
	var req struct {
		Token string `json:"token"`
	}
	if e := c.Bind(&req); e != nil {
		return utils.BadRequest(c, "Tautan tidak valid.")
	}
	v, e := h.Service.Exchange(c.Request().Context(), c.Param("site"), req.Token)
	if e != nil {
		return portalError(c, e)
	}
	return utils.Success(c, "Berhasil masuk", v)
}
func bearer(c echo.Context) string {
	parts := strings.Fields(c.Request().Header.Get("Authorization"))
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return parts[1]
	}
	return ""
}
func (h *Handler) identity(c echo.Context) (*service.Identity, error) {
	return h.Service.Authenticate(c.Request().Context(), c.Param("site"), bearer(c))
}
func page(c echo.Context, key string) int {
	p, _ := strconv.Atoi(c.QueryParam(key))
	if p < 1 {
		return 1
	}
	return p
}
func (h *Handler) Overview(c echo.Context) error {
	v, e := h.identity(c)
	if e != nil {
		return portalError(c, e)
	}
	data, e := h.Service.Overview(c.Request().Context(), v, page(c, "subscription_page"), page(c, "invoice_page"))
	if e != nil {
		return portalError(c, e)
	}
	return utils.Success(c, "Langganan dan invoice", data)
}
func (h *Handler) Logout(c echo.Context) error {
	if _, e := h.identity(c); e != nil {
		return portalError(c, e)
	}
	if e := h.Service.Logout(c.Request().Context(), bearer(c)); e != nil {
		return portalError(c, e)
	}
	return utils.Success(c, "Berhasil keluar", nil)
}
func (h *Handler) owned(c echo.Context) (*service.Identity, int, error) {
	v, e := h.identity(c)
	if e != nil {
		return nil, 0, e
	}
	id, e := strconv.Atoi(c.Param("id"))
	if e != nil || id < 1 {
		return nil, 0, utils.ErrBadRequest
	}
	e = h.Service.OwnInvoice(c.Request().Context(), v, id)
	return v, id, e
}
func (h *Handler) InvoiceDetail(c echo.Context) error {
	v, id, e := h.owned(c)
	if e != nil {
		return portalError(c, e)
	}
	inv, e := h.Invoice.GetDetail(id, v.BusinessID)
	if e != nil {
		return portalError(c, e)
	}
	return utils.Success(c, "Detail invoice", map[string]interface{}{"invoice": inv.InvoiceListItem, "items": inv.Items})
}
func (h *Handler) Pay(c echo.Context) error {
	v, id, e := h.owned(c)
	if e != nil {
		return portalError(c, e)
	}
	link, e := h.Checkout.CreatePortalLink(c.Request().Context(), id, v.BusinessID)
	if e != nil {
		return portalError(c, e)
	}
	return utils.Success(c, "Lanjutkan pembayaran", link)
}
