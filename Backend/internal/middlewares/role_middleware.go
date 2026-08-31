package middlewares

import (
	"pay-langgan/internal/models"
	"pay-langgan/internal/utils"

	"github.com/labstack/echo/v4"
)

func RequireRoles(roles ...models.Role) echo.MiddlewareFunc {
	allowed := make(map[models.Role]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			role := models.Role(GetRole(c))
			if _, ok := allowed[role]; !ok {
				return utils.Forbidden(c, "you do not have permission to access this resource")
			}
			return next(c)
		}
	}
}

func RequireTenantUser(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if !models.IsTenantRole(GetRole(c)) || GetBusinessID(c) == "" {
			return utils.Forbidden(c, "this resource is only available to business users")
		}
		return next(c)
	}
}
