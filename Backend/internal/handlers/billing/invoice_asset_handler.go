package billing

import (
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/labstack/echo/v4"
	"pay-langgan/internal/config"
)

var logoFilename = regexp.MustCompile(`^[a-f0-9]{64}\.(png|jpg)$`)

// Public, immutable logo files only; no metadata, auth, listing or arbitrary paths.
func InvoiceLogoAsset(cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		name := c.Param("file")
		if !logoFilename.MatchString(name) {
			return echo.ErrNotFound
		}
		file, err := os.Open(filepath.Join(cfg.InvoiceAssetDir, name))
		if err != nil {
			return echo.ErrNotFound
		}
		defer file.Close()
		info, err := file.Stat()
		if err != nil || !info.Mode().IsRegular() || info.Size() > 1024*1024 {
			return echo.ErrNotFound
		}
		header := make([]byte, 512)
		n, _ := file.Read(header)
		kind := http.DetectContentType(header[:n])
		if kind != "image/png" && kind != "image/jpeg" {
			return echo.ErrNotFound
		}
		_, _ = file.Seek(0, 0)
		c.Response().Header().Set("Content-Type", kind)
		c.Response().Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		c.Response().Header().Set("X-Content-Type-Options", "nosniff")
		http.ServeContent(c.Response(), c.Request(), name, info.ModTime(), file)
		return nil
	}
}
