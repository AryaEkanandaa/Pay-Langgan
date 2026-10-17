package billing

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"pay-langgan/internal/models"
)

func TestInvoicePDF(t *testing.T) {
	inv := &models.InvoiceDetailResponse{InvoiceListItem: models.InvoiceListItem{
		InvoiceNumber: "INV-PDF-TEST", CustomerName: "Indira Utari", PlanName: "Enterprise Bulanan",
		Amount: 218500, Status: "paid", CreatedAt: time.Now(),
	}, Items: []models.InvoiceItemResponse{
		{Description: "Enterprise Bulanan", Quantity: 1, Subtotal: 200000},
		{Description: "Laporan Bisnis Lanjutan", Quantity: 1, Subtotal: 30000},
		{Description: "Coupon 1STTIME", Quantity: 1, Subtotal: -11500},
	}}
	doc, err := RenderInvoicePDF(inv, "Kelana Workspace", json.RawMessage(`{"accent_color":"#3654ff","footer_note":"Terima kasih."}`))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(doc, []byte("%PDF-")) || !bytes.Contains(doc, []byte("%%EOF")) {
		t.Fatal("invalid PDF")
	}
	// Long descriptions and many items must paginate without failing.
	for range 100 {
		inv.Items = append(inv.Items, models.InvoiceItemResponse{Description: strings.Repeat("Tambahan ruang rapat ", 30), Quantity: 1, Subtotal: 30000})
	}
	if _, err := RenderInvoicePDF(inv, "Kelana", json.RawMessage(`{}`)); err != nil {
		t.Fatal(err)
	}
	if _, err := RenderInvoicePDF(inv, "Kelana", json.RawMessage(`invalid`)); err == nil {
		t.Fatal("invalid appearance accepted")
	}
}
