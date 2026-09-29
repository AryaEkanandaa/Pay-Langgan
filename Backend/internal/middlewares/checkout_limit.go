package middlewares

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"pay-langgan/internal/utils"
)

type CheckoutLimiter interface {
	AllowCheckoutRequest(context.Context, string, int) (bool, error)
}

func CheckoutRateLimit(store CheckoutLimiter, write bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip, _, err := net.SplitHostPort(c.Request().RemoteAddr)
			if err != nil {
				ip = c.Request().RemoteAddr
			}
			// Do not trust client-supplied forwarding headers. Reverse proxies must
			// apply an additional per-client limit at the edge.
			if net.ParseIP(ip) == nil {
				ip = "unknown-peer"
			}
			kind, ipLimit, tokenLimit := "read", 60, 30
			if write {
				kind, ipLimit, tokenLimit = "session", 10, 5
			}
			ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
			defer cancel()
			for _, bucket := range []struct {
				key   string
				limit int
			}{
				{kind + ":ip:" + ip, ipLimit}, {kind + ":token:" + c.Param("token"), tokenLimit},
			} {
				allowed, err := store.AllowCheckoutRequest(ctx, bucket.key, bucket.limit)
				if err != nil {
					return utils.Error(c, http.StatusServiceUnavailable, "checkout temporarily unavailable")
				}
				if !allowed {
					c.Response().Header().Set("Retry-After", "60")
					return utils.Error(c, http.StatusTooManyRequests, "too many checkout requests; try again shortly")
				}
			}
			return next(c)
		}
	}
}
