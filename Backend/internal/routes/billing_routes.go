package routes

import (
	"pay-langgan/internal/middlewares"
	"pay-langgan/internal/models"

	"github.com/labstack/echo/v4"
)

func registerBillingRoutes(g *echo.Group, h *AllHandlers) {
	read := []echo.MiddlewareFunc{
		middlewares.RequireTenantUser,
		middlewares.RequireRoles(models.RoleAdmin, models.RoleSales, models.RoleFinance),
	}
	write := []echo.MiddlewareFunc{
		middlewares.RequireTenantUser,
		middlewares.RequireRoles(models.RoleAdmin, models.RoleFinance),
	}

	g.GET("/invoices", h.Invoice.List, read...)
	g.GET("/invoices/:id", h.Invoice.GetByID, read...)
	if h.InvoiceEmail != nil {
		g.POST("/invoice-appearance/preview", h.InvoiceEmail.AppearancePreview, read...)
		g.GET("/email-history", h.InvoiceEmail.History, read...)
		g.GET("/email-history/:source/:id/preview", h.InvoiceEmail.HistoryPreview, read...)
		g.POST("/email-history/:source/:id/retry", h.InvoiceEmail.RetryHistory, write...)
		g.GET("/email-reminders", h.InvoiceEmail.ReminderSettings, read...)
		g.PUT("/email-reminders", h.InvoiceEmail.SaveReminderSettings, write...)
		g.GET("/invoices/:id/email", h.InvoiceEmail.Status, read...)
		g.GET("/invoices/:id/email-preview", h.InvoiceEmail.Preview, read...)
		g.POST("/invoices/:id/email", h.InvoiceEmail.Queue, write...)
	}
	g.POST("/invoices/:id/mark-paid", h.Invoice.MarkPaid, write...)
	g.POST("/invoices/:id/checkout-links", h.Checkout.CreateLink, write...)
	g.GET("/payments", h.Checkout.History, write...)
	g.GET("/customers/:id/payment-methods", h.Checkout.Methods, write...)
	g.GET("/subscriptions/:id/autopay", h.Checkout.GetAutopay, write...)
	g.PUT("/subscriptions/:id/autopay", h.Checkout.SetAutopay, write...)
}
