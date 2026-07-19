package routes

import "github.com/labstack/echo/v4"

func registerCatalogRoutes(g *echo.Group, h *AllHandlers) {
	g.GET("/services", h.Service.List)
	g.GET("/services/:id", h.Service.GetByID)
	g.POST("/services", h.Service.Create)
	g.PUT("/services/:id", h.Service.Update)
	g.DELETE("/services/:id", h.Service.Delete)

	g.GET("/products", h.Product.List)
	g.GET("/products/:id", h.Product.GetByID)
	g.POST("/products", h.Product.Create)
	g.PUT("/products/:id", h.Product.Update)
	g.DELETE("/products/:id", h.Product.Delete)

	g.GET("/plans", h.Plan.List)
	g.GET("/plans/:id", h.Plan.GetByID)
	g.POST("/plans", h.Plan.Create)
	g.PUT("/plans/:id", h.Plan.Update)
	g.DELETE("/plans/:id", h.Plan.Delete)

	g.GET("/add-ons", h.AddOn.List)
	g.GET("/add-ons/:id", h.AddOn.GetByID)
	g.POST("/add-ons", h.AddOn.Create)
	g.PUT("/add-ons/:id", h.AddOn.Update)
	g.DELETE("/add-ons/:id", h.AddOn.Delete)
}
