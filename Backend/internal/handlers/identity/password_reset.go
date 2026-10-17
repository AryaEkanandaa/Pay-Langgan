package identity

import (
	"errors"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
	"net"
	"net/http"
	"pay-langgan/internal/middlewares"
	"pay-langgan/internal/utils"
	"time"
	"unicode/utf8"
)

func (h *AuthHandler) PasswordResetLimit() echo.MiddlewareFunc {
	return middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(middleware.RateLimiterMemoryStoreConfig{Rate: rate.Limit(1.0 / 60), Burst: 5, ExpiresIn: 5 * time.Minute}),
		IdentifierExtractor: func(c echo.Context) (string, error) {
			host, _, err := net.SplitHostPort(c.Request().RemoteAddr)
			if err != nil {
				return c.Request().RemoteAddr, nil
			}
			return host, nil
		},
		DenyHandler: func(c echo.Context, _ string, _ error) error {
			return c.JSON(http.StatusTooManyRequests, map[string]interface{}{"success": false, "message": "Terlalu banyak permintaan. Tunggu sebentar lalu coba lagi."})
		},
	})
}

func (h *AuthHandler) ForgotPassword(c echo.Context) error {
	c.Response().Header().Set("Cache-Control", "no-store")
	c.Request().Body = http.MaxBytesReader(c.Response(), c.Request().Body, 4096)
	started := time.Now()
	var request struct {
		Email string `json:"email"`
	}
	if err := c.Bind(&request); err != nil {
		return utils.BadRequest(c, "Masukkan alamat email yang valid.")
	}
	err := h.authService.RequestPasswordReset(c.Request().Context(), request.Email)
	// Both existing and unknown accounts follow the same response floor; no SMTP in this request.
	timer := time.NewTimer(max(0, 200*time.Millisecond-time.Since(started)))
	defer timer.Stop()
	select {
	case <-timer.C:
	case <-c.Request().Context().Done():
		return nil
	}
	if errors.Is(err, utils.ErrBadRequest) {
		return utils.BadRequest(c, "Masukkan alamat email yang valid.")
	}
	if err != nil {
		return utils.InternalError(c, "Permintaan belum dapat diproses. Silakan coba lagi nanti.")
	}
	return utils.Success(c, "Jika email terdaftar, tautan untuk mengatur ulang password akan dikirim. Periksa inbox dan folder spam.", nil)
}

func (h *AuthHandler) ResetPassword(c echo.Context) error {
	c.Response().Header().Set("Cache-Control", "no-store")
	c.Request().Body = http.MaxBytesReader(c.Response(), c.Request().Body, 4096)
	var request struct {
		Token        string `json:"token"`
		Password     string `json:"password"`
		Confirmation string `json:"password_confirmation"`
	}
	if err := c.Bind(&request); err != nil {
		return utils.BadRequest(c, "Data password baru tidak valid.")
	}
	if request.Password != request.Confirmation || utf8.RuneCountInString(request.Password) < 8 || len(request.Password) > 72 {
		return utils.BadRequest(c, "Password minimal 8 karakter, maksimal 72 byte, dan konfirmasi harus sama.")
	}
	err := h.authService.ResetPassword(c.Request().Context(), request.Token, request.Password, request.Confirmation)
	if errors.Is(err, utils.ErrBadRequest) {
		return utils.BadRequest(c, "Tautan tidak valid, sudah digunakan, atau kedaluwarsa. Silakan minta tautan baru.")
	}
	if err != nil {
		return utils.InternalError(c, "Password belum dapat diperbarui. Silakan coba lagi.")
	}
	return utils.Success(c, "Password berhasil diperbarui. Silakan masuk dengan password baru.", nil)
}

func (h *AuthHandler) RequireCurrentSession(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		version, _ := c.Get("password_version").(int)
		err := h.authService.VerifySession(c.Request().Context(), middlewares.GetUserID(c), version)
		if errors.Is(err, utils.ErrUnauthorized) {
			return utils.Unauthorized(c, "Sesi telah berakhir. Silakan masuk kembali.")
		}
		if err != nil {
			return utils.InternalError(c, "Sesi belum dapat diperiksa.")
		}
		return next(c)
	}
}
