package main

import (
	"fmt"
	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"pay-langgan/internal/config"
	"pay-langgan/internal/database"
	billinghttp "pay-langgan/internal/handlers/billing"
	cataloghttp "pay-langgan/internal/handlers/catalog"
	couponhttp "pay-langgan/internal/handlers/coupon"
	customerhttp "pay-langgan/internal/handlers/customer"
	identityhttp "pay-langgan/internal/handlers/identity"
	revenuehttp "pay-langgan/internal/handlers/revenue"
	subscriptionhttp "pay-langgan/internal/handlers/subscription"
	"pay-langgan/internal/repositories/audit"
	billingrepo "pay-langgan/internal/repositories/billing"
	catalogrepo "pay-langgan/internal/repositories/catalog"
	couponrepo "pay-langgan/internal/repositories/coupon"
	customerrepo "pay-langgan/internal/repositories/customer"
	identityrepo "pay-langgan/internal/repositories/identity"
	revenuerepo "pay-langgan/internal/repositories/revenue"
	subscriptionrepo "pay-langgan/internal/repositories/subscription"
	"pay-langgan/internal/routes"
	billingsvc "pay-langgan/internal/services/billing"
	catalogsvc "pay-langgan/internal/services/catalog"
	couponsvc "pay-langgan/internal/services/coupon"
	customersvc "pay-langgan/internal/services/customer"
	identitysvc "pay-langgan/internal/services/identity"
	revenuesvc "pay-langgan/internal/services/revenue"
	subscriptionsvc "pay-langgan/internal/services/subscription"
)

func main() {
	cfg := config.Load()

	db, err := database.New(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	authRepo := identityrepo.NewAuthRepository(db)
	authService := identitysvc.NewAuthService(authRepo, cfg)
	authHandler := identityhttp.NewAuthHandler(authService)

	businessRepo := identityrepo.NewBusinessRepository(db)
	userRepo := identityrepo.NewUserRepository(db)
	serviceRepo := catalogrepo.NewServiceRepository(db)
	productRepo := catalogrepo.NewProductRepository(db)
	planRepo := catalogrepo.NewPlanRepository(db)
	addOnRepo := catalogrepo.NewAddOnRepository(db)
	couponRepo := couponrepo.NewCouponRepository(db)
	customerRepo := customerrepo.NewCustomerRepository(db)
	subRepo := subscriptionrepo.NewSubscriptionRepository(db)
	subAddOnRepo := subscriptionrepo.NewSubscriptionAddOnRepository(db)
	subCpnRepo := subscriptionrepo.NewSubscriptionCouponRepository(db)
	auditLogRepo := audit.NewAuditLogRepository(db)
	invoiceRepo := billingrepo.NewInvoiceRepository(db)
	businessService := identitysvc.NewBusinessService(businessRepo)
	userService := identitysvc.NewUserService(userRepo)
	serviceService := catalogsvc.NewServiceService(serviceRepo)
	productService := catalogsvc.NewProductService(productRepo, serviceRepo)
	planService := catalogsvc.NewPlanService(planRepo, productRepo)
	addOnService := catalogsvc.NewAddOnService(addOnRepo, productRepo)
	couponService := couponsvc.NewCouponService(couponRepo)
	customerService := customersvc.NewCustomerService(customerRepo)
	subscriptionPricingService := subscriptionsvc.NewSubscriptionPricingService(planRepo, addOnRepo, couponRepo)
	subscriptionService := subscriptionsvc.NewSubscriptionService(db, subRepo, subAddOnRepo, subCpnRepo, auditLogRepo, planRepo, productRepo, serviceRepo, addOnRepo, couponRepo, customerRepo, invoiceRepo)
	dashboardRepo := revenuerepo.NewDashboardRepository(db)
	dashboardService := revenuesvc.NewDashboardService(dashboardRepo)
	invoiceService := billingsvc.NewInvoiceService(db, invoiceRepo)

	businessHandler := identityhttp.NewBusinessHandler(businessService)
	userHandler := identityhttp.NewUserHandler(userService)
	serviceHandler := cataloghttp.NewServiceHandler(serviceService)
	productHandler := cataloghttp.NewProductHandler(productService)
	planHandler := cataloghttp.NewPlanHandler(planService)
	addOnHandler := cataloghttp.NewAddOnHandler(addOnService)
	couponHandler := couponhttp.NewCouponHandler(couponService)
	customerHandler := customerhttp.NewCustomerHandler(customerService)
	subscriptionHandler := subscriptionhttp.NewSubscriptionHandler(subscriptionService, subscriptionPricingService)
	dashboardHandler := revenuehttp.NewDashboardHandler(dashboardService)
	invoiceHandler := billinghttp.NewInvoiceHandler(invoiceService)

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]interface{}{
			"success": true,
			"message": fmt.Sprintf("%s is running", cfg.AppName),
			"data": map[string]interface{}{
				"app_name": cfg.AppName,
				"env":      cfg.AppEnv,
			},
		})
	})

	routes.RegisterRoutes(e, &routes.AllHandlers{
		Auth:         authHandler,
		Business:     businessHandler,
		User:         userHandler,
		Service:      serviceHandler,
		Product:      productHandler,
		Plan:         planHandler,
		AddOn:        addOnHandler,
		Coupon:       couponHandler,
		Customer:     customerHandler,
		Subscription: subscriptionHandler,
		Invoice:      invoiceHandler,
		Dashboard:    dashboardHandler,
	}, cfg.JWTSecret)

	addr := fmt.Sprintf(":%s", cfg.AppPort)
	log.Printf("%s starting on %s", cfg.AppName, addr)
	e.Logger.Fatal(e.Start(addr))
}
