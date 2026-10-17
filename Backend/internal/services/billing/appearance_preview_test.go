package billing

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"pay-langgan/internal/config"
)

func TestAppearancePreviewUsesDeliveryTemplates(t *testing.T) {
	now := time.Date(2026, 10, 8, 6, 0, 0, 0, time.UTC)
	cfg := &config.Config{InvoiceAssetDir: t.TempDir()}
	raw := json.RawMessage(`{"accent_color":"#6345ff","address":"Alamat contoh","footer_note":"Catatan contoh","email_subject":"Tagihan {{invoice_number}} - {{business_name}}","email_message":"Halo {{customer_name}}, total {{total}}, jatuh tempo {{due_date}}","logo_url":"https://temporary.trycloudflare.com/logo.png"}`)
	prepared, _, err := PrepareInvoiceEmailAppearance(raw, cfg)
	if err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{"invoice_issued", "payment_succeeded"} {
		got, err := PreviewInvoiceAppearance("Bisnis Contoh", raw, kind, cfg, now)
		if err != nil {
			t.Fatal(err)
		}
		inv := appearanceSampleInvoice(now)
		var subject, body string
		if kind == "payment_succeeded" {
			inv.Status = "paid"
			inv.PaidAt = &now
			subject, body, _, err = renderPaymentConfirmation(inv, "Bisnis Contoh", prepared, "#invoice-contoh")
		} else {
			subject, body, _, err = RenderInvoiceEmail(inv, "Bisnis Contoh", prepared, "#invoice-contoh")
		}
		if err != nil || got.Subject != subject || got.HTML != body {
			t.Fatalf("%s differs from delivery template: %v", kind, err)
		}
		if got.LogoWarning == "" || strings.Contains(got.HTML, "<img") {
			t.Fatal("temporary logo must be omitted with explanation")
		}
		for _, value := range []string{"Alamat contoh", "Catatan contoh", "Rp 218.500", "Add-on Contoh", "-Rp 11.500", "#6345ff"} {
			if !strings.Contains(got.HTML, value) {
				t.Fatal("missing preview field", value)
			}
		}
	}
	if _, err := PreviewInvoiceAppearance("Bisnis", raw, "invalid", cfg, now); err == nil {
		t.Fatal("unsupported preview kind accepted")
	}
}
