package billing

import (
	"errors"
	"github.com/labstack/echo/v4"
	"pay-langgan/internal/middlewares"
	"pay-langgan/internal/utils"
	"strconv"
)

func (h *EmailHandler) History(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	items, total, err := h.Invoice.CustomerEmailHistory(c.Request().Context(), middlewares.GetBusinessID(c), c.QueryParam("status"), c.QueryParam("kind"), c.QueryParam("search"), page)
	if err != nil {
		return historyError(c, err)
	}
	c.Response().Header().Set("Cache-Control", "no-store")
	return utils.Success(c, "Riwayat email", map[string]interface{}{"items": items, "total": total, "page": page, "limit": 20, "mode": h.Config.EmailDeliveryMode})
}
func (h *EmailHandler) RetryHistory(c echo.Context) error {
	if h.Config.EmailDeliveryMode == "disabled" {
		return utils.BadRequest(c, "Pengiriman email belum aktif.")
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id < 1 {
		return utils.BadRequest(c, "ID email tidak valid.")
	}
	if err = h.Invoice.RetryCustomerEmail(c.Request().Context(), middlewares.GetBusinessID(c), c.Param("source"), id); err != nil {
		return historyError(c, err)
	}
	return utils.Success(c, "Email masuk antrean kembali", nil)
}
func (h *EmailHandler) HistoryPreview(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id < 1 {
		return utils.BadRequest(c, "ID email tidak valid.")
	}
	subject, body, err := h.Invoice.CustomerEmailPreview(c.Request().Context(), middlewares.GetBusinessID(c), c.Param("source"), id)
	if errors.Is(err, utils.ErrConflict) {
		return c.JSON(409, map[string]interface{}{"success": false, "message": "Preview belum tersedia. Email lama sebelum fitur riwayat tidak memiliki salinan HTML; email baru tersedia setelah diproses."})
	}
	if err != nil {
		return historyError(c, err)
	}
	c.Response().Header().Set("Cache-Control", "no-store")
	return utils.Success(c, "Preview email", map[string]string{"subject": subject, "html": body})
}
func (h *EmailHandler) ReminderSettings(c echo.Context) error {
	c.Response().Header().Set("Cache-Control", "no-store")
	p, err := h.Invoice.ReminderPreferences(c.Request().Context(), middlewares.GetBusinessID(c))
	if err != nil {
		return historyError(c, err)
	}
	return utils.Success(c, "Pengingat email", p)
}
func (h *EmailHandler) SaveReminderSettings(c echo.Context) error {
	var req struct {
		Enabled *bool `json:"enabled"`
	}
	if err := c.Bind(&req); err != nil || req.Enabled == nil {
		return utils.BadRequest(c, "Pilihan pengingat tidak valid.")
	}
	p, err := h.Invoice.SetReminderPreferences(c.Request().Context(), middlewares.GetBusinessID(c), *req.Enabled)
	if err != nil {
		return historyError(c, err)
	}
	return utils.Success(c, "Pengaturan pengingat disimpan", p)
}
func historyError(c echo.Context, err error) error {
	if errors.Is(err, utils.ErrBadRequest) {
		return utils.BadRequest(c, "Filter atau ID email tidak valid.")
	}
	if errors.Is(err, utils.ErrNotFound) {
		return utils.NotFound(c, "Email tidak ditemukan.")
	}
	if errors.Is(err, utils.ErrConflict) {
		return c.JSON(409, map[string]interface{}{"success": false, "message": "Email tidak lagi berstatus gagal. Muat ulang riwayat email."})
	}
	return utils.InternalError(c, "Gagal memproses riwayat email.")
}
