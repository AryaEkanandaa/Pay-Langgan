package routes

import (
	"pay-langgan/internal/middlewares"
	"pay-langgan/internal/models"

	"github.com/labstack/echo/v4"
)

func registerCatalogRoutes(g *echo.Group, h *AllHandlers) {
	read := []echo.MiddlewareFunc{middlewares.RequireTenantUser, middlewares.RequireRoles(models.RoleAdmin, models.RoleSales, models.RoleFinance)}
	write := []echo.MiddlewareFunc{middlewares.RequireTenantUser, middlewares.RequireRoles(models.RoleAdmin, models.RoleSales)}

	g.GET("/services", h.Service.List, read...)
	g.GET("/services/:id", h.Service.GetByID, read...)
	g.POST("/services", h.Service.Create, write...)
	g.PUT("/services/:id", h.Service.Update, write...)
	g.DELETE("/services/:id", h.Service.Delete, write...)

	g.GET("/products", h.Product.List, read...)
	g.GET("/products/:id", h.Product.GetByID, read...)
	g.POST("/products", h.Product.Create, write...)
	g.PUT("/products/:id", h.Product.Update, write...)
	g.DELETE("/products/:id", h.Product.Delete, write...)

	g.GET("/plans", h.Plan.List, read...)
	g.GET("/plans/:id", h.Plan.GetByID, read...)
	g.POST("/plans", h.Plan.Create, write...)
	g.PUT("/plans/:id", h.Plan.Update, write...)
	g.DELETE("/plans/:id", h.Plan.Delete, write...)

	g.GET("/add-ons", h.AddOn.List, read...)
	g.GET("/add-ons/:id", h.AddOn.GetByID, read...)
	g.POST("/add-ons", h.AddOn.Create, write...)
	g.PUT("/add-ons/:id", h.AddOn.Update, write...)
	g.DELETE("/add-ons/:id", h.AddOn.Delete, write...)
}
