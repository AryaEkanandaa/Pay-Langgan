package routes

import (
	"pay-langgan/internal/middlewares"
	"pay-langgan/internal/models"

	"github.com/labstack/echo/v4"
)

func registerCustomerRoutes(g *echo.Group, h *AllHandlers) {
	read := []echo.MiddlewareFunc{middlewares.RequireTenantUser, middlewares.RequireRoles(models.RoleAdmin, models.RoleSales, models.RoleFinance)}
	write := []echo.MiddlewareFunc{middlewares.RequireTenantUser, middlewares.RequireRoles(models.RoleAdmin, models.RoleSales)}

	g.GET("/customers", h.Customer.List, read...)
	g.GET("/customers/:id", h.Customer.GetByID, read...)
	g.POST("/customers", h.Customer.Create, write...)
	g.PUT("/customers/:id", h.Customer.Update, write...)
	g.DELETE("/customers/:id", h.Customer.Delete, write...)
}
