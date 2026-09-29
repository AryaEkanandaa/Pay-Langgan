package middlewares

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

type testCheckoutLimiter struct {
	keys []string
	deny bool
	err  error
}

func (l *testCheckoutLimiter) AllowCheckoutRequest(_ context.Context, key string, _ int) (bool, error) {
	l.keys = append(l.keys, key)
	return !l.deny, l.err
}

func TestCheckoutLimitForwardedHeadersAndFailures(t *testing.T) {
	for _, test := range []struct {
		name   string
		deny   bool
		err    error
		status int
	}{
		{"allowed", false, nil, 200}, {"limited", true, nil, 429}, {"storage error", false, errors.New("offline"), 503},
	} {
		t.Run(test.name, func(t *testing.T) {
			store := &testCheckoutLimiter{deny: test.deny, err: test.err}
			e := echo.New()
			e.POST("/checkout/:token/session", func(c echo.Context) error { return c.NoContent(200) }, CheckoutRateLimit(store, true))
			r := httptest.NewRequest(http.MethodPost, "/checkout/test/session", nil)
			r.RemoteAddr = "192.0.2.3:1234"
			r.Header.Set("X-Forwarded-For", "8.8.8.8")
			w := httptest.NewRecorder()
			e.ServeHTTP(w, r)
			if w.Code != test.status || store.keys[0] != "session:ip:192.0.2.3" {
				t.Fatalf("status=%d keys=%v", w.Code, store.keys)
			}
			if test.deny && w.Header().Get("Retry-After") != "60" {
				t.Fatal("missing retry hint")
			}
		})
	}
}
