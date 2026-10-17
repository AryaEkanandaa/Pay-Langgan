package billing

import (
	"encoding/json"
	"errors"
	"github.com/labstack/echo/v4"
	"html"
	"mime"
	"net/http"
	"pay-langgan/internal/config"
	"pay-langgan/internal/middlewares"
	svc "pay-langgan/internal/services/billing"
	payment "pay-langgan/internal/services/payment"
	"pay-langgan/internal/utils"
	"strconv"
	"strings"
	"time"
)

type EmailHandler struct {
	Invoice  *svc.InvoiceService
	Checkout *payment.CheckoutService
	Config   *config.Config
}

func (h *EmailHandler) AppearancePreview(c echo.Context) error {
	c.Response().Header().Set("Cache-Control", "no-store")
	c.Request().Body = http.MaxBytesReader(c.Response(), c.Request().Body, 2<<20)
	var request struct {
		Settings json.RawMessage `json:"settings"`
		Kind     string          `json:"kind"`
	}
	if err := c.Bind(&request); err != nil || len(request.Settings) == 0 || string(request.Settings) == "null" {
		return utils.BadRequest(c, "Pengaturan preview tidak valid.")
	}
	if request.Kind == "" {
		request.Kind = "invoice_issued"
	}
	if request.Kind != "invoice_issued" && request.Kind != "payment_succeeded" {
		return utils.BadRequest(c, "Jenis preview tidak valid.")
	}
	var name string
	var ignored json.RawMessage
	if err := h.Invoice.EmailSettings(c.Request().Context(), middlewares.GetBusinessID(c), &name, &ignored); err != nil {
		return emailError(c, err)
	}
	preview, err := svc.PreviewInvoiceAppearance(name, request.Settings, request.Kind, h.Config, time.Now().UTC())
	if err != nil {
		return utils.BadRequest(c, "Preview belum dapat dibuat. Periksa pengaturan invoice.")
	}
	return utils.Success(c, "Invoice appearance preview", preview)
}

func (h *EmailHandler) Status(c echo.Context) error {
	id, e := strconv.Atoi(c.Param("id"))
	if e != nil || id < 1 {
		return utils.BadRequest(c, "invalid invoice id")
	}
	result, e := h.Invoice.EmailStatus(c.Request().Context(), id, middlewares.GetBusinessID(c))
	if e != nil {
		return emailError(c, e)
	}
	c.Response().Header().Set("Cache-Control", "no-store")
	return utils.Success(c, "Email status", map[string]interface{}{"delivery": result, "mode": h.Config.EmailDeliveryMode})
}
func (h *EmailHandler) Queue(c echo.Context) error {
	if h.Config.EmailDeliveryMode == "disabled" {
		return utils.BadRequest(c, "Invoice email delivery is disabled")
	}
	id, e := strconv.Atoi(c.Param("id"))
	if e != nil || id < 1 {
		return utils.BadRequest(c, "invalid invoice id")
	}
	result, e := h.Invoice.QueueEmail(c.Request().Context(), id, middlewares.GetBusinessID(c))
	if e != nil {
		return emailError(c, e)
	}
	return utils.Success(c, "Invoice email queued", map[string]interface{}{"delivery": result, "mode": h.Config.EmailDeliveryMode})
}
func emailError(c echo.Context, e error) error {
	if errors.Is(e, utils.ErrNotFound) {
		return utils.NotFound(c, "invoice not found")
	}
	if errors.Is(e, utils.ErrBadRequest) {
		return utils.BadRequest(c, "Alamat email pelanggan belum diisi atau tidak valid.")
	}
	return utils.InternalError(c, "Gagal memproses email invoice.")
}
func (h *EmailHandler) Preview(c echo.Context) error {
	c.Response().Header().Set("Cache-Control", "no-store")
	id, e := strconv.Atoi(c.Param("id"))
	if e != nil || id < 1 {
		return utils.BadRequest(c, "invalid invoice id")
	}
	biz := middlewares.GetBusinessID(c)
	inv, e := h.Invoice.GetDetail(id, biz)
	if e != nil {
		return emailError(c, e)
	}
	var name string
	var raw json.RawMessage
	e = h.Invoice.EmailPreviewSettings(c.Request().Context(), id, biz, &name, &inv.CustomerName, &raw)
	if e != nil {
		return emailError(c, e)
	}
	prepared, warning, e := svc.PrepareInvoiceEmailAppearance(raw, h.Config)
	if e != nil {
		return emailError(c, e)
	}
	recipient := ""
	if inv.CustomerEmail != nil {
		recipient = *inv.CustomerEmail
	}
	var subject, body string
	if h.Config.EmailRendererURL != "" {
		subject, body, _, e = svc.RenderReactInvoicePreview(c.Request().Context(), h.Config, inv, name, recipient, prepared)
	} else {
		subject, body, _, e = svc.RenderInvoiceEmail(inv, name, prepared, "")
	}
	if e != nil {
		return emailError(c, e)
	}
	return utils.Success(c, "Email preview", map[string]string{"subject": subject, "html": body, "logo_warning": warning})
}
func (h *EmailHandler) Public(c echo.Context) error {
	id, biz, raw, e := h.Invoice.PublicEmailInvoice(c.Request().Context(), c.Param("token"))
	if e != nil {
		return utils.NotFound(c, "invoice link unavailable")
	}
	inv, e := h.Invoice.GetDetail(id, biz)
	if e != nil {
		return emailError(c, e)
	}
	var name string
	var ignored json.RawMessage
	if e = h.Invoice.EmailSettings(c.Request().Context(), biz, &name, &ignored); e != nil {
		return emailError(c, e)
	}
	prepared, _, e := svc.PrepareInvoiceEmailAppearance(raw, h.Config)
	if e != nil {
		return emailError(c, e)
	}
	token := html.EscapeString(c.Param("token"))
	checkoutPath := ""
	if inv.Status == "pending" {
		checkoutPath = "/api/v1/invoice-emails/" + token + "/checkout"
	}
	downloadPath := "/api/v1/invoice-emails/" + token + "/download"
	var body string
	if h.Config.EmailRendererURL != "" {
		_, body, _, e = svc.RenderReactPublicInvoice(c.Request().Context(), h.Config, inv, name,
			func() string {
				if inv.CustomerEmail != nil {
					return *inv.CustomerEmail
				}
				return ""
			}(), prepared, checkoutPath, downloadPath)
	} else {
		_, body, _, e = svc.RenderInvoiceEmail(inv, name, prepared, "")
	}
	if e != nil {
		return emailError(c, e)
	}
	if h.Config.EmailRendererURL == "" {
		actions := "<section style=\"margin-top:26px\">"
		if checkoutPath != "" {
			actions += `<form method="post" action="` + checkoutPath + `" style="margin:0 0 12px"><button type="submit" style="background:#3654ff;color:white;padding:13px 20px;border:0;border-radius:9px;cursor:pointer;font:700 13px Arial,sans-serif">Lanjut ke pembayaran</button></form>`
		}
		actions += `<a href="` + downloadPath + `" style="display:inline-block;color:#2442a3;padding:12px 20px;border:1px solid #cbd5e1;border-radius:9px;text-decoration:none;font:700 13px Arial,sans-serif">Unduh PDF</a></section>`
		body = strings.Replace(body, "</main>", actions+"</main>", 1)
	}
	c.Response().Header().Set("Cache-Control", "no-store")
	c.Response().Header().Set("Referrer-Policy", "no-referrer")
	c.Response().Header().Set("Content-Security-Policy", "default-src 'none'; img-src https: data:; style-src 'unsafe-inline'; form-action 'self'; frame-ancestors 'none'")
	return c.HTML(http.StatusOK, body)
}

// Download applies the same token expiry and invoice ownership checks as Public.
func (h *EmailHandler) Download(c echo.Context) error {
	c.Response().Header().Set("Cache-Control", "no-store")
	c.Response().Header().Set("Referrer-Policy", "no-referrer")
	id, biz, raw, err := h.Invoice.PublicEmailInvoice(c.Request().Context(), c.Param("token"))
	if err != nil {
		return utils.NotFound(c, "invoice link unavailable")
	}
	inv, err := h.Invoice.GetDetail(id, biz)
	if err != nil {
		return emailError(c, err)
	}
	var name string
	var ignored json.RawMessage
	if err = h.Invoice.EmailSettings(c.Request().Context(), biz, &name, &ignored); err != nil {
		return emailError(c, err)
	}
	document, err := svc.RenderInvoicePDF(inv, name, raw)
	if err != nil {
		return utils.InternalError(c, "PDF invoice belum dapat dibuat. Silakan coba lagi.")
	}
	c.Response().Header().Set("X-Content-Type-Options", "nosniff")
	filename := strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, inv.InvoiceNumber)
	c.Response().Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": filename + ".pdf"}))
	return c.Blob(http.StatusOK, "application/pdf", document)
}
func (h *EmailHandler) Pay(c echo.Context) error {
	id, biz, _, e := h.Invoice.PublicEmailInvoice(c.Request().Context(), c.Param("token"))
	if e != nil {
		return utils.NotFound(c, "invoice link unavailable")
	}
	link, e := h.Checkout.CreateLink(c.Request().Context(), id, biz)
	if e != nil {
		return c.String(http.StatusConflict, "Pembayaran belum dapat dibuka. Jika pembayaran sudah dimulai, lanjutkan dari halaman pembayaran sebelumnya atau hubungi bisnis.")
	}
	c.Response().Header().Set("Cache-Control", "no-store")
	return c.Redirect(http.StatusSeeOther, link.CheckoutURL)
}
