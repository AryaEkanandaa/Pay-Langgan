package middlewares

import (
	"strings"

	"pay-langgan/internal/utils"
	"github.com/labstack/echo/v4"
)

type Skipper func(c echo.Context) bool

func JWTAuth(secret string, skipper Skipper) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if skipper(c) {
				return next(c)
			}

			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return utils.Unauthorized(c, "missing authorization header")
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				return utils.Unauthorized(c, "invalid authorization header format")
			}

			claims, err := utils.ParseToken(secret, parts[1])
			if err != nil {
				return utils.Unauthorized(c, "invalid or expired token")
			}

			c.Set("user_id", claims.UserID)
			c.Set("business_id", claims.BusinessID)
			c.Set("email", claims.Email)
			c.Set("role", claims.Role)

			return next(c)
		}
	}
}

func GetUserID(c echo.Context) int {
	if uid, ok := c.Get("user_id").(int); ok {
		return uid
	}
	return 0
}

func GetBusinessID(c echo.Context) string {
	if bid, ok := c.Get("business_id").(string); ok {
		return bid
	}
	return ""
}

func GetRole(c echo.Context) string {
	if role, ok := c.Get("role").(string); ok {
		return role
	}
	return ""
}

func GetEmail(c echo.Context) string {
	if email, ok := c.Get("email").(string); ok {
		return email
	}
	return ""
}
