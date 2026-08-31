package routes

import (
	"pay-langgan/internal/middlewares"
	"pay-langgan/internal/models"

	"github.com/labstack/echo/v4"
)

func registerCouponRoutes(g *echo.Group, h *AllHandlers) {
	read := []echo.MiddlewareFunc{middlewares.RequireTenantUser, middlewares.RequireRoles(models.RoleAdmin, models.RoleSales, models.RoleFinance)}
	write := []echo.MiddlewareFunc{middlewares.RequireTenantUser, middlewares.RequireRoles(models.RoleAdmin, models.RoleSales)}

	g.GET("/coupons", h.Coupon.List, read...)
	g.GET("/coupons/:id", h.Coupon.GetByID, read...)
	g.POST("/coupons", h.Coupon.Create, write...)
	g.PUT("/coupons/:id", h.Coupon.Update, write...)
	g.DELETE("/coupons/:id", h.Coupon.Delete, write...)
}
