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
	g.POST("/invoices/:id/mark-paid", h.Invoice.MarkPaid, write...)
	g.POST("/invoices/:id/checkout-links", h.Checkout.CreateLink, write...)
	g.GET("/payments", h.Checkout.History, write...)
	g.GET("/customers/:id/payment-methods", h.Checkout.Methods, write...)
	g.GET("/subscriptions/:id/autopay", h.Checkout.GetAutopay, write...)
	g.PUT("/subscriptions/:id/autopay", h.Checkout.SetAutopay, write...)
}
