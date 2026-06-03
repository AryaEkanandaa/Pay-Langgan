package main

import (
	"fmt"
	"log"

	"pay-langgan/internal/config"
	"pay-langgan/internal/database"
	"pay-langgan/internal/handlers"
	"pay-langgan/internal/repositories"
	"pay-langgan/internal/routes"
	"pay-langgan/internal/services"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	cfg := config.Load()

	db, err := database.New(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	authRepo := repositories.NewAuthRepository(db)
	authService := services.NewAuthService(authRepo, cfg)
	authHandler := handlers.NewAuthHandler(authService)

	businessRepo := repositories.NewBusinessRepository(db)
	serviceRepo := repositories.NewServiceRepository(db)
	productRepo := repositories.NewProductRepository(db)
	planRepo := repositories.NewPlanRepository(db)
	addOnRepo := repositories.NewAddOnRepository(db)
	couponRepo := repositories.NewCouponRepository(db)
	customerRepo := repositories.NewCustomerRepository(db)
	subRepo := repositories.NewSubscriptionRepository(db)
	subAddOnRepo := repositories.NewSubscriptionAddOnRepository(db)
	subCpnRepo := repositories.NewSubscriptionCouponRepository(db)
	auditLogRepo := repositories.NewAuditLogRepository(db)
	businessService := services.NewBusinessService(businessRepo)
	serviceService := services.NewServiceService(serviceRepo)
	productService := services.NewProductService(productRepo, serviceRepo)
	planService := services.NewPlanService(planRepo, productRepo)
	addOnService := services.NewAddOnService(addOnRepo, productRepo)
	couponService := services.NewCouponService(couponRepo)
	customerService := services.NewCustomerService(customerRepo)
	subscriptionPricingService := services.NewSubscriptionPricingService(planRepo, addOnRepo, couponRepo)
	subscriptionService := services.NewSubscriptionService(db, subRepo, subAddOnRepo, subCpnRepo, auditLogRepo, planRepo, productRepo, serviceRepo, addOnRepo, couponRepo, customerRepo)

	businessHandler := handlers.NewBusinessHandler(businessService)
	serviceHandler := handlers.NewServiceHandler(serviceService)
	productHandler := handlers.NewProductHandler(productService)
	planHandler := handlers.NewPlanHandler(planService)
	addOnHandler := handlers.NewAddOnHandler(addOnService)
	couponHandler := handlers.NewCouponHandler(couponService)
	customerHandler := handlers.NewCustomerHandler(customerService)
	subscriptionHandler := handlers.NewSubscriptionHandler(subscriptionService, subscriptionPricingService)

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
		Service:      serviceHandler,
		Product:      productHandler,
		Plan:         planHandler,
		AddOn:        addOnHandler,
		Coupon:       couponHandler,
		Customer:     customerHandler,
		Subscription: subscriptionHandler,
	}, cfg.JWTSecret)

	addr := fmt.Sprintf(":%s", cfg.AppPort)
	log.Printf("%s starting on %s", cfg.AppName, addr)
	e.Logger.Fatal(e.Start(addr))
}
