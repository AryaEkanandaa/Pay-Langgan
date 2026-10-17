package billing

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pay-langgan/internal/config"
	"pay-langgan/internal/models"
)

func TestInvoiceEmailPublicLogoStorage(t *testing.T) {
	var imageData bytes.Buffer
	if err := png.Encode(&imageData, image.NewRGBA(image.Rect(0, 0, 40, 80))); err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(map[string]string{"logo_url": "data:image/png;base64," + base64.StdEncoding.EncodeToString(imageData.Bytes()), "footer_note": "Saved footer"})
	cfg := &config.Config{InvoiceAssetDir: t.TempDir()}
	prepared, warning, err := PrepareInvoiceEmailAppearance(raw, cfg)
	if err != nil || warning == "" {
		t.Fatalf("unconfigured storage: %s %v", warning, err)
	}
	_, html, _, err := RenderInvoiceEmail(&models.InvoiceDetailResponse{}, "Kelana", prepared, "")
	if err != nil || strings.Contains(html, "<img") || !strings.Contains(html, "Kelana") {
		t.Fatal("missing-URL fallback", err)
	}
	cfg.PublicAssetURL = "https://assets.paylanggan.example/public-assets/invoice-logos"
	prepared, warning, err = PrepareInvoiceEmailAppearance(raw, cfg)
	if err != nil || warning != "" {
		t.Fatalf("public logo: %s %v", warning, err)
	}
	var a EmailAppearance
	_ = json.Unmarshal(prepared, &a)
	if !strings.HasPrefix(a.Logo, cfg.PublicAssetURL+"/") || !strings.HasSuffix(a.Logo, ".png") || a.LogoWidth != 35 || a.LogoHeight != 70 {
		t.Fatalf("logo URL/size: %+v", a)
	}
	data, err := os.ReadFile(filepath.Join(cfg.InvoiceAssetDir, filepath.Base(a.Logo)))
	if err != nil || !bytes.Equal(data, imageData.Bytes()) {
		t.Fatal("stored bytes differ", err)
	}
	_, html, _, err = RenderInvoiceEmail(&models.InvoiceDetailResponse{}, "Kelana", prepared, "")
	if err != nil || !strings.Contains(html, `width="35" height="70" style="display:block`) || strings.Contains(html, "data:image") || !strings.Contains(html, "Saved footer") {
		t.Fatal("email rendering", err)
	}
	// A snapshot containing the resolved URL is unchanged by future base changes.
	cfg.PublicAssetURL = "https://new.paylanggan.example/logos"
	next, _, err := PrepareInvoiceEmailAppearance(prepared, cfg)
	if err != nil || !bytes.Equal(prepared, next) {
		t.Fatal("preview changed a saved logo URL", err)
	}
}

func TestInvoiceEmailRejectsUnsafeLogoURLs(t *testing.T) {
	for _, value := range []string{"", "data:image/png;base64,AA==", "/logo.png", "http://public.example/logo.png", "https://localhost/logo.png", "https://127.0.0.1/logo.png", "https://192.168.1.2/logo.jpg", "https://demo.trycloudflare.com/logo.png", "https://cdn.example/logo.svg", "https://cdn.example/logo.webp", "https://user:pass@cdn.example/logo.png"} {
		if ValidEmailLogoURL(value) {
			t.Errorf("unsafe logo accepted: %s", value)
		}
	}
	for _, value := range []string{"https://cdn.example/logo.png", "https://cdn.example/logo.jpg", "https://cdn.example/logo.jpeg"} {
		if !ValidEmailLogoURL(value) {
			t.Errorf("valid logo rejected: %s", value)
		}
	}
}
