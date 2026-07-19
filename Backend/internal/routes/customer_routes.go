package routes

import "github.com/labstack/echo/v4"

func registerCustomerRoutes(g *echo.Group, h *AllHandlers) {
	g.GET("/customers", h.Customer.List)
	g.GET("/customers/:id", h.Customer.GetByID)
	g.POST("/customers", h.Customer.Create)
	g.PUT("/customers/:id", h.Customer.Update)
	g.DELETE("/customers/:id", h.Customer.Delete)
}
