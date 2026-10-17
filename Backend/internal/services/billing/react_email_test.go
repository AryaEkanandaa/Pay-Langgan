package billing

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"pay-langgan/internal/config"
	"pay-langgan/internal/models"
	"strings"
	"testing"
	"time"
)

func TestReactEmailBridgeMapsInvoiceAndPreservesAppearance(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/render" || r.Method != http.MethodPost {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var payload struct {
			Type  string                 `json:"type"`
			Props map[string]interface{} `json:"props"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Error(err)
		}
		if payload.Type != "invoice-issued" || payload.Props["subjectOverride"] != "Tagihan INV-1 - Kelana" {
			t.Errorf("incorrect template or subject: %#v", payload)
		}
		invoice := payload.Props["invoice"].(map[string]interface{})
		if invoice["total"] != float64(218500) || len(invoice["items"].([]interface{})) != 2 {
			t.Errorf("invoice data was lost: %#v", invoice)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"subject":"Tagihan INV-1 - Kelana","html":"<p>Siap dibayar</p>"}`))
	}))
	defer server.Close()
	now := time.Date(2026, 10, 15, 8, 0, 0, 0, time.FixedZone("WITA", 8*3600))
	inv := &models.InvoiceDetailResponse{
		InvoiceListItem: models.InvoiceListItem{InvoiceNumber: "INV-1", CustomerName: "Arya", Amount: 218500, Status: "pending", CreatedAt: now},
		Items:           []models.InvoiceItemResponse{{Description: "Paket", Quantity: 1, UnitPrice: 230000, Subtotal: 230000}, {Description: "Kupon", Quantity: 1, UnitPrice: -11500, Subtotal: -11500}},
	}
	cfg := &config.Config{EmailRendererURL: server.URL}
	raw := json.RawMessage(`{"email_subject":"Tagihan {{invoice_number}} - {{business_name}}","email_message":"Halo {{customer_name}}","reply_to":"bantuan@kelana.example","accent_color":"#3B4EFF"}`)
	subject, html, reply, err := renderQueuedInvoice(context.Background(), cfg, "invoice-issued", inv, "Kelana", "arya@example.com", raw, "https://paylanggan.example/invoice/1")
	if err != nil || subject != "Tagihan INV-1 - Kelana" || html != "<p>Siap dibayar</p>" || reply != "bantuan@kelana.example" {
		t.Fatalf("bridge result: %q %q %q %v", subject, html, reply, err)
	}
}

func TestReactEmailBridgeDoesNotSilentlySendAnotherTemplate(t *testing.T) {
	cfg := &config.Config{EmailRendererURL: "http://127.0.0.1:1"}
	inv := &models.InvoiceDetailResponse{}
	_, _, _, err := renderQueuedInvoice(context.Background(), cfg, "invoice-issued", inv, "Kelana", "arya@example.com", json.RawMessage(`{}`), "")
	if err == nil || !strings.Contains(err.Error(), "renderer unavailable") {
		t.Fatalf("expected renderer failure, got %v", err)
	}
	if err = validateEmailRendererURL("https://external.example"); err == nil {
		t.Fatal("external renderer URL accepted")
	}
}
