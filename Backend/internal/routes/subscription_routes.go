package routes

import (
	"pay-langgan/internal/middlewares"
	"pay-langgan/internal/models"

	"github.com/labstack/echo/v4"
)

func registerSubscriptionRoutes(g *echo.Group, h *AllHandlers) {
	read := []echo.MiddlewareFunc{middlewares.RequireTenantUser, middlewares.RequireRoles(models.RoleAdmin, models.RoleSales, models.RoleFinance)}
	write := []echo.MiddlewareFunc{middlewares.RequireTenantUser, middlewares.RequireRoles(models.RoleAdmin, models.RoleSales)}

	g.POST("/subscriptions/preview", h.Subscription.Preview, write...)
	g.GET("/subscriptions", h.Subscription.List, read...)
	g.GET("/subscriptions/:id", h.Subscription.GetByID, read...)
	g.POST("/subscriptions", h.Subscription.Create, write...)
	g.POST("/subscriptions/:id/cancel", h.Subscription.Cancel, write...)
	g.POST("/subscriptions/:id/pause", h.Subscription.Pause, write...)
	g.POST("/subscriptions/:id/resume", h.Subscription.Resume, write...)
	g.POST("/subscriptions/:id/add-ons", h.Subscription.AddAddOn, write...)
	g.DELETE("/subscriptions/:id/add-ons/:add_on_id", h.Subscription.RemoveAddOn, write...)
	g.POST("/subscriptions/:id/coupons", h.Subscription.ApplyCoupon, write...)
	g.DELETE("/subscriptions/:id/coupons/:coupon_id", h.Subscription.RemoveCoupon, write...)
}
