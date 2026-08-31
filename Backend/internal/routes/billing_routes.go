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
}
