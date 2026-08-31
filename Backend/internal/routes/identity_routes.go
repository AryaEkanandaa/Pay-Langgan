package routes

import (
	"pay-langgan/internal/middlewares"
	"pay-langgan/internal/models"

	"github.com/labstack/echo/v4"
)

func registerIdentityRoutes(g *echo.Group, h *AllHandlers) {
	g.GET("/me", h.Auth.Me)

	g.GET("/businesses/me", h.Business.GetMyBusiness,
		middlewares.RequireTenantUser,
		middlewares.RequireRoles(models.RoleAdmin, models.RoleSales, models.RoleFinance),
	)
	g.PUT("/businesses/me", h.Business.UpdateMyBusiness,
		middlewares.RequireTenantUser,
		middlewares.RequireRoles(models.RoleAdmin),
	)

	g.GET("/users", h.User.List,
		middlewares.RequireTenantUser,
		middlewares.RequireRoles(models.RoleAdmin),
	)
	g.POST("/users", h.User.CreateStaff,
		middlewares.RequireTenantUser,
		middlewares.RequireRoles(models.RoleAdmin),
	)
}
