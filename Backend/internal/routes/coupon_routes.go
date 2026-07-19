package routes

import "github.com/labstack/echo/v4"

func registerCouponRoutes(g *echo.Group, h *AllHandlers) {
	g.GET("/coupons", h.Coupon.List)
	g.GET("/coupons/:id", h.Coupon.GetByID)
	g.POST("/coupons", h.Coupon.Create)
	g.PUT("/coupons/:id", h.Coupon.Update)
	g.DELETE("/coupons/:id", h.Coupon.Delete)
}
