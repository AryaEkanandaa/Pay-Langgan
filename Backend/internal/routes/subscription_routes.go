package routes

import "github.com/labstack/echo/v4"

func registerSubscriptionRoutes(g *echo.Group, h *AllHandlers) {
	g.POST("/subscriptions/preview", h.Subscription.Preview)
	g.GET("/subscriptions", h.Subscription.List)
	g.GET("/subscriptions/:id", h.Subscription.GetByID)
	g.POST("/subscriptions", h.Subscription.Create)
	g.POST("/subscriptions/:id/cancel", h.Subscription.Cancel)
	g.POST("/subscriptions/:id/pause", h.Subscription.Pause)
	g.POST("/subscriptions/:id/resume", h.Subscription.Resume)
	g.POST("/subscriptions/:id/add-ons", h.Subscription.AddAddOn)
	g.DELETE("/subscriptions/:id/add-ons/:add_on_id", h.Subscription.RemoveAddOn)
	g.POST("/subscriptions/:id/coupons", h.Subscription.ApplyCoupon)
	g.DELETE("/subscriptions/:id/coupons/:coupon_id", h.Subscription.RemoveCoupon)
}
