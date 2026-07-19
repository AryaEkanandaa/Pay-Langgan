package routes

import "github.com/labstack/echo/v4"

func registerIdentityRoutes(g *echo.Group, h *AllHandlers) {
	g.GET("/me", h.Auth.Me)

	g.GET("/businesses/me", h.Business.GetMyBusiness)
	g.PUT("/businesses/me", h.Business.UpdateMyBusiness)
}
