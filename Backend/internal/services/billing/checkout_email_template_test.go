package billing

import (
	"encoding/json"
	"pay-langgan/internal/models"
	"strings"
	"testing"
	"time"
)

func TestCheckoutConfirmationTemplates(t *testing.T) {
	now := time.Date(2026, 10, 8, 1, 0, 0, 0, time.UTC)
	inv := &models.InvoiceDetailResponse{InvoiceListItem: models.InvoiceListItem{Status: "pending", InvoiceNumber: "INV-TEST", Amount: 100000, CustomerName: "<script>buyer</script>", PaidAt: &now, PeriodStart: &now, PeriodEnd: &now}}
	if _, _, _, err := renderPaymentConfirmation(inv, "Kelana", json.RawMessage(`{}`), ""); err == nil {
		t.Fatal("pending invoice rendered as paid")
	}
	inv.Status = "paid"
	subject, body, reply, err := renderPaymentConfirmation(inv, "Kelana", json.RawMessage(`{"accent_color":"#123456","footer_note":"Hubungi tim Kelana","reply_to":"help@example.com","email_message":"Bayar sekarang"}`), "https://example.com/invoice")
	if err != nil || !strings.Contains(subject, "Pembayaran berhasil") || !strings.Contains(body, "Lunas") || !strings.Contains(body, "#123456") || strings.Contains(body, "Bayar sekarang") || strings.Contains(body, "<script>buyer") || reply != "help@example.com" {
		t.Fatalf("receipt %s %s %s %v", subject, body, reply, err)
	}
	inv.Amount = 0
	_, body, _, err = renderPaymentConfirmation(inv, "Kelana", json.RawMessage(`{}`), "")
	if err != nil || !strings.Contains(body, "tidak ada dana") {
		t.Fatal("zero checkout represented as card charge")
	}
	snapshot, _ := json.Marshal(trialEmailSnapshot{Plan: "Pro", Cycle: "monthly", EndsAt: now, Amount: 125000, Items: []models.SubscriptionPreviewItem{{Name: "Storage", Quantity: 1, Subtotal: 25000}}})
	subject, body, _, err = renderTrialConfirmation("Kelana", "Buyer", json.RawMessage(`{"logo_url":"data:image/png;base64,AAA","footer_note":"Hubungi Kelana"}`), snapshot)
	for _, text := range []string{"Trial berhasil dimulai", "Rp 0", "Rp 125.000", "Storage", "08/10/2026", "perlu dibayar secara manual", "Hubungi Kelana"} {
		if !strings.Contains(body, text) {
			t.Fatalf("trial missing %s", text)
		}
	}
	if err != nil || strings.Contains(body, "<img") || !strings.Contains(subject, "Trial dimulai") {
		t.Fatalf("trial %v", err)
	}
	snapshot, _ = json.Marshal(trialEmailSnapshot{Plan: "Pro", Cycle: "monthly", EndsAt: now, Amount: 125000, AutoDebitEnabled: true, PaymentMethod: "midtrans"})
	_, body, _, err = renderTrialConfirmation("Kelana", "Buyer", json.RawMessage(`{}`), snapshot)
	if err != nil || !strings.Contains(body, "kartu yang terdaftar di Midtrans akan didebit otomatis") || strings.Contains(body, "perlu dibayar secara manual") {
		t.Fatalf("Midtrans trial autopay disclosure missing: %v", err)
	}
}
