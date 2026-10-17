package billing

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"pay-langgan/internal/config"
)

func stableHTTPSURL(value string) bool {
	u, err := url.Parse(value)
	if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.Fragment != "" {
		return false
	}
	host := strings.ToLower(strings.TrimSuffix(u.Hostname(), "."))
	if host == "localhost" || !strings.Contains(host, ".") || strings.HasSuffix(host, ".localhost") || strings.HasSuffix(host, ".local") || host == "trycloudflare.com" || strings.HasSuffix(host, ".trycloudflare.com") {
		return false
	}
	if ip := net.ParseIP(host); ip != nil {
		return false
	}
	return true
}

func ValidEmailLogoURL(value string) bool {
	if !stableHTTPSURL(value) {
		return false
	}
	u, _ := url.Parse(value)
	switch strings.ToLower(path.Ext(u.Path)) {
	case ".png", ".jpg", ".jpeg":
		return true
	}
	return false
}

// Both previews and the worker use this function. Uploaded images remain in
// business metadata for print compatibility, but are never embedded in emails.
func PrepareInvoiceEmailAppearance(raw json.RawMessage, cfg *config.Config) (json.RawMessage, string, error) {
	var settings map[string]json.RawMessage
	if err := json.Unmarshal(raw, &settings); err != nil {
		return nil, "", err
	}
	if settings == nil {
		settings = make(map[string]json.RawMessage)
	}
	var logo string
	_ = json.Unmarshal(settings["logo_url"], &logo)
	if logo == "" {
		return raw, "", nil
	}
	warning := ""
	width, height := 150, 70
	if strings.HasPrefix(logo, "data:") {
		filename, w, h, err := storeInvoiceLogo(logo, cfg.InvoiceAssetDir)
		if err != nil {
			settings["logo_url"] = json.RawMessage(`""`)
			result, marshalErr := json.Marshal(settings)
			return result, "Logo belum dapat disiapkan sebagai PNG/JPG publik. Email tetap menampilkan nama bisnis.", marshalErr
		}
		width, height = w, h
		base := strings.TrimRight(cfg.PublicAssetURL, "/")
		if base == "" || !stableHTTPSURL(base) {
			logo = ""
			warning = "Logo sudah disimpan, tetapi belum ada alamat gambar HTTPS permanen. Email menampilkan nama bisnis sampai penyimpanan publik dihubungkan."
		} else {
			u, _ := url.Parse(base)
			if u.RawQuery != "" {
				return nil, "", fmt.Errorf("PUBLIC_ASSET_URL cannot contain query parameters")
			}
			logo = base + "/" + filename
		}
	} else if !ValidEmailLogoURL(logo) {
		logo = ""
		warning = "Logo email harus memakai URL HTTPS permanen berformat PNG/JPG yang dapat dibuka tanpa login."
	} else {
		// Preserve the exact URL and dimensions captured by the delivery job.
		return raw, "", nil
	}
	settings["logo_url"], _ = json.Marshal(logo)
	settings["logo_width"], _ = json.Marshal(width)
	settings["logo_height"], _ = json.Marshal(height)
	result, err := json.Marshal(settings)
	return result, warning, err
}

func storeInvoiceLogo(value, dir string) (string, int, int, error) {
	if len(value) > 1_400_000 {
		return "", 0, 0, fmt.Errorf("logo exceeds 1 MB")
	}
	parts := strings.SplitN(value, ",", 2)
	if len(parts) != 2 || (parts[0] != "data:image/png;base64" && parts[0] != "data:image/jpeg;base64") {
		return "", 0, 0, fmt.Errorf("logo must be PNG or JPG")
	}
	data, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil || len(data) > 1024*1024 {
		return "", 0, 0, fmt.Errorf("invalid logo data")
	}
	info, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || info.Width < 1 || info.Height < 1 || info.Width > 4096 || info.Height > 4096 {
		return "", 0, 0, fmt.Errorf("invalid logo dimensions")
	}
	if (format != "png" && format != "jpeg") || parts[0] != "data:image/"+format+";base64" {
		return "", 0, 0, fmt.Errorf("logo format mismatch")
	}
	if _, _, err := image.Decode(bytes.NewReader(data)); err != nil {
		return "", 0, 0, fmt.Errorf("invalid logo image")
	}
	ext := ".png"
	if format == "jpeg" {
		ext = ".jpg"
	}
	filename := fmt.Sprintf("%x%s", sha256.Sum256(data), ext)
	if dir == "" {
		dir = "./storage/invoice-logos"
	}
	if err = os.MkdirAll(dir, 0755); err != nil {
		return "", 0, 0, err
	}
	file, err := os.CreateTemp(dir, ".logo-*")
	if err != nil {
		return "", 0, 0, err
	}
	temporary := file.Name()
	defer os.Remove(temporary)
	_, writeErr := file.Write(data)
	closeErr := file.Close()
	if writeErr != nil {
		return "", 0, 0, writeErr
	}
	if closeErr != nil {
		return "", 0, 0, closeErr
	}
	target := filepath.Join(dir, filename)
	if err = os.Rename(temporary, target); err != nil {
		// On Windows another concurrent request may already have published it.
		existing, readErr := os.ReadFile(target)
		if readErr != nil || !bytes.Equal(existing, data) {
			return "", 0, 0, err
		}
	}
	width, height := 150, info.Height*150/info.Width
	if height > 70 {
		height = 70
		width = info.Width * 70 / info.Height
	}
	return filename, max(1, width), max(1, height), nil
}
