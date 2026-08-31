package routes

import (
	"net/http"

	"pay-langgan/internal/handlers/billing"
	"pay-langgan/internal/handlers/catalog"
	"pay-langgan/internal/handlers/coupon"
	"pay-langgan/internal/handlers/customer"
	"pay-langgan/internal/handlers/identity"
	"pay-langgan/internal/handlers/revenue"
	"pay-langgan/internal/handlers/subscription"
	"pay-langgan/internal/middlewares"
	"pay-langgan/internal/models"

	"github.com/labstack/echo/v4"
)

type SkipperFunc func(c echo.Context) bool

type AllHandlers struct {
	Auth         *identity.AuthHandler
	Business     *identity.BusinessHandler
	User         *identity.UserHandler
	Service      *catalog.ServiceHandler
	Product      *catalog.ProductHandler
	Plan         *catalog.PlanHandler
	AddOn        *catalog.AddOnHandler
	Coupon       *coupon.CouponHandler
	Customer     *customer.CustomerHandler
	Subscription *subscription.SubscriptionHandler
	Invoice      *billing.InvoiceHandler
	Dashboard    *revenue.DashboardHandler
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

	registerIdentityRoutes(protected, h)
	registerCatalogRoutes(protected, h)
	registerCouponRoutes(protected, h)
	registerCustomerRoutes(protected, h)
	registerSubscriptionRoutes(protected, h)
	registerBillingRoutes(protected, h)
	protected.GET("/dashboard/summary", h.Dashboard.Summary,
		middlewares.RequireTenantUser,
		middlewares.RequireRoles(models.RoleAdmin, models.RoleSales, models.RoleFinance),
	)
}
