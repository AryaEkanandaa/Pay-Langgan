package routes

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"pay-langgan/internal/handlers/billing"
	"pay-langgan/internal/handlers/catalog"
	"pay-langgan/internal/handlers/coupon"
	"pay-langgan/internal/handlers/customer"
	"pay-langgan/internal/handlers/identity"
	paymenthttp "pay-langgan/internal/handlers/payment"
	platformhttp "pay-langgan/internal/handlers/platformbilling"
	"pay-langgan/internal/handlers/revenue"
	"pay-langgan/internal/handlers/subscription"
	"pay-langgan/internal/middlewares"
	"pay-langgan/internal/models"

	"github.com/labstack/echo/v4"
)

type SkipperFunc func(c echo.Context) bool

type AllHandlers struct {
	Auth            *identity.AuthHandler
	Business        *identity.BusinessHandler
	User            *identity.UserHandler
	Service         *catalog.ServiceHandler
	Product         *catalog.ProductHandler
	Plan            *catalog.PlanHandler
	AddOn           *catalog.AddOnHandler
	Coupon          *coupon.CouponHandler
	Customer        *customer.CustomerHandler
	Subscription    *subscription.SubscriptionHandler
	Invoice         *billing.InvoiceHandler
	Checkout        *paymenthttp.CheckoutHandler
	Merchant        *paymenthttp.MerchantIntegrationHandler
	BusinessGateway *paymenthttp.BusinessGatewayHandler
	PlatformBilling *platformhttp.Handler
	Dashboard       *revenue.DashboardHandler
}

func RegisterRoutes(e *echo.Echo, h *AllHandlers, jwtSecret string) {
	api := e.Group("/api/v1")

	api.POST("/signup", h.Auth.Signup)
	api.POST("/login", h.Auth.Login)
	api.GET("/checkout/:token", h.Checkout.Get, h.Checkout.RateLimit(false))
	api.POST("/checkout/:token/session", h.Checkout.Start, h.Checkout.RateLimit(true))
	api.POST("/integrations/orders", h.Merchant.CreateOrder)
	api.GET("/integrations/orders/:reference", h.Merchant.GetOrder)
	api.POST("/embed/sites/:site/orders", h.Merchant.CreateEmbedOrder)
	api.POST("/webhooks/xendit/payment-sessions", func(c echo.Context) error {
		c.Request().Body = http.MaxBytesReader(c.Response(), c.Request().Body, 1<<20)
		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return c.NoContent(http.StatusBadRequest)
		}
		var event struct {
			Data struct {
				ReferenceID string `json:"reference_id"`
			} `json:"data"`
		}
		if json.Unmarshal(body, &event) != nil {
			return c.NoContent(http.StatusBadRequest)
		}
		c.Request().Body = io.NopCloser(bytes.NewReader(body))
		if strings.HasPrefix(event.Data.ReferenceID, "platform_") {
			return h.PlatformBilling.Webhook(c)
		}
		return h.Checkout.XenditWebhook(c)
	})
	api.POST("/webhooks/business-gateways/:webhookID", h.Checkout.BusinessGatewayWebhook)
	api.POST("/webhooks/xendit/platform-payment-sessions", h.PlatformBilling.Webhook)

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
	merchantKeys := protected.Group("/integrations/keys", middlewares.RequireTenantUser, middlewares.RequireRoles(models.RoleAdmin))
	merchantKeys.POST("", h.Merchant.CreateKey)
	merchantKeys.GET("", h.Merchant.ListKeys)
	merchantKeys.DELETE("/:id", h.Merchant.RevokeKey)
	businessGateway := protected.Group("/settings/payment-gateway", middlewares.RequireTenantUser, middlewares.RequireRoles(models.RoleAdmin))
	businessGateway.GET("", h.BusinessGateway.Get)
	businessGateway.PUT("", h.BusinessGateway.Save)
	businessGateway.DELETE("", h.BusinessGateway.Delete)
	embedSites := protected.Group("/integrations/embed-sites", middlewares.RequireTenantUser, middlewares.RequireRoles(models.RoleAdmin))
	embedSites.POST("", h.Merchant.CreateEmbedSite)
	embedSites.GET("", h.Merchant.ListEmbedSites)
	embedSites.DELETE("/:id", h.Merchant.RevokeEmbedSite)
	platform := protected.Group("/platform", middlewares.RequireTenantUser, middlewares.RequireRoles(models.RoleAdmin))
	platform.GET("/plan", h.PlatformBilling.Plan)
	platform.GET("/subscription", h.PlatformBilling.Status)
	platform.GET("/invoices", h.PlatformBilling.Invoices)
	platform.POST("/subscription/checkout", h.PlatformBilling.Start)
	platform.GET("/subscription/checkout", h.PlatformBilling.Resume)
	platform.POST("/subscription/cancel", h.PlatformBilling.Cancel)
	protected.GET("/dashboard/summary", h.Dashboard.Summary,
		middlewares.RequireTenantUser,
		middlewares.RequireRoles(models.RoleAdmin, models.RoleSales, models.RoleFinance),
	)
	protected.GET("/platform/dashboard/summary", h.Dashboard.PlatformSummary,
		middlewares.RequireRoles(models.RoleSuperAdmin),
	)
}
