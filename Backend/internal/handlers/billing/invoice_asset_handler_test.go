package billing

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"pay-langgan/internal/config"
)

func TestInvoiceLogoPublicPNGAndJPEG(t *testing.T) {
	dir := t.TempDir()
	e := echo.New()
	e.GET("/public-assets/invoice-logos/:file", InvoiceLogoAsset(&config.Config{InvoiceAssetDir: dir}))
	for _, format := range []string{"png", "jpg"} {
		var data bytes.Buffer
		img := image.NewRGBA(image.Rect(0, 0, 2, 2))
		if format == "png" {
			_ = png.Encode(&data, img)
		} else {
			_ = jpeg.Encode(&data, img, nil)
		}
		name := strings.Repeat("a", 64) + "." + format
		if err := os.WriteFile(filepath.Join(dir, name), data.Bytes(), 0600); err != nil {
			t.Fatal(err)
		}
		res := httptest.NewRecorder()
		e.ServeHTTP(res, httptest.NewRequest("GET", "/public-assets/invoice-logos/"+name, nil))
		expected := "image/png"
		if format == "jpg" {
			expected = "image/jpeg"
		}
		if res.Code != 200 || res.Header().Get("Content-Type") != expected || !bytes.Equal(res.Body.Bytes(), data.Bytes()) {
			t.Fatalf("anonymous image: %d %s", res.Code, res.Header().Get("Content-Type"))
		}
		if res.Header().Get("X-Content-Type-Options") != "nosniff" {
			t.Fatal("missing nosniff")
		}
	}
	for _, name := range []string{"secret.env", "logo.svg", strings.Repeat("b", 64) + ".png", "%2e%2e%2f.env"} {
		res := httptest.NewRecorder()
		e.ServeHTTP(res, httptest.NewRequest("GET", "/public-assets/invoice-logos/"+name, nil))
		if res.Code != 404 {
			t.Errorf("non-logo accessible: %s %d", name, res.Code)
		}
	}
}
