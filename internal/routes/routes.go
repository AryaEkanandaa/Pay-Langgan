package routes

import (
	"net/http"

	"pay-langgan/internal/handlers"
	"pay-langgan/internal/middlewares"

	"github.com/labstack/echo/v4"
)

type SkipperFunc func(c echo.Context) bool

type AllHandlers struct {
	Auth         *handlers.AuthHandler
	Business     *handlers.BusinessHandler
	Service      *handlers.ServiceHandler
	Product      *handlers.ProductHandler
	Plan         *handlers.PlanHandler
	AddOn        *handlers.AddOnHandler
	Coupon       *handlers.CouponHandler
	Customer     *handlers.CustomerHandler
	Subscription *handlers.SubscriptionHandler
}

func RegisterRoutes(e *echo.Echo, h *AllHandlers, jwtSecret string) {
	api := e.Group("/api/v1")

	api.POST("/signup", h.Auth.Signup)
	api.POST("/login", h.Auth.Login)

	protected := api.Group("")
	protected.Use(middlewares.JWTAuth(jwtSecret, func(c echo.Context) bool {
		path := c.Path()
		method := c.Request().Method
		if method == http.MethodPost && (path == "/api/v1/signup" || path == "/api/v1/login") {
			return true
		}
		return false
	}))

	protected.GET("/me", h.Auth.Me)

	protected.GET("/businesses/me", h.Business.GetMyBusiness)
	protected.PUT("/businesses/me", h.Business.UpdateMyBusiness)

	protected.GET("/services", h.Service.List)
	protected.GET("/services/:id", h.Service.GetByID)
	protected.POST("/services", h.Service.Create)
	protected.PUT("/services/:id", h.Service.Update)
	protected.DELETE("/services/:id", h.Service.Delete)

	protected.GET("/products", h.Product.List)
	protected.GET("/products/:id", h.Product.GetByID)
	protected.POST("/products", h.Product.Create)
	protected.PUT("/products/:id", h.Product.Update)
	protected.DELETE("/products/:id", h.Product.Delete)

	protected.GET("/plans", h.Plan.List)
	protected.GET("/plans/:id", h.Plan.GetByID)
	protected.POST("/plans", h.Plan.Create)
	protected.PUT("/plans/:id", h.Plan.Update)
	protected.DELETE("/plans/:id", h.Plan.Delete)

	protected.GET("/add-ons", h.AddOn.List)
	protected.GET("/add-ons/:id", h.AddOn.GetByID)
	protected.POST("/add-ons", h.AddOn.Create)
	protected.PUT("/add-ons/:id", h.AddOn.Update)
	protected.DELETE("/add-ons/:id", h.AddOn.Delete)

	protected.GET("/coupons", h.Coupon.List)
	protected.GET("/coupons/:id", h.Coupon.GetByID)
	protected.POST("/coupons", h.Coupon.Create)
	protected.PUT("/coupons/:id", h.Coupon.Update)
	protected.DELETE("/coupons/:id", h.Coupon.Delete)

	protected.GET("/customers", h.Customer.List)
	protected.GET("/customers/:id", h.Customer.GetByID)
	protected.POST("/customers", h.Customer.Create)
	protected.PUT("/customers/:id", h.Customer.Update)
	protected.DELETE("/customers/:id", h.Customer.Delete)

	protected.POST("/subscriptions/preview", h.Subscription.Preview)
	protected.GET("/subscriptions", h.Subscription.List)
	protected.GET("/subscriptions/:id", h.Subscription.GetByID)
	protected.POST("/subscriptions", h.Subscription.Create)
	protected.POST("/subscriptions/:id/cancel", h.Subscription.Cancel)
	protected.POST("/subscriptions/:id/pause", h.Subscription.Pause)
	protected.POST("/subscriptions/:id/resume", h.Subscription.Resume)
	protected.POST("/subscriptions/:id/add-ons", h.Subscription.AddAddOn)
	protected.DELETE("/subscriptions/:id/add-ons/:add_on_id", h.Subscription.RemoveAddOn)
	protected.POST("/subscriptions/:id/coupons", h.Subscription.ApplyCoupon)
	protected.DELETE("/subscriptions/:id/coupons/:coupon_id", h.Subscription.RemoveCoupon)
}
