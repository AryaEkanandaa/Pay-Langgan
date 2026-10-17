package main

import (
	"context"
	"encoding/json"
	"log"
	"net/mail"
	"os"
	"pay-langgan/internal/config"
	"pay-langgan/internal/models"
	"pay-langgan/internal/services/billing"
	"pay-langgan/internal/services/mailer"
	"time"
)

// Sends one clearly marked sample to an explicitly supplied test address.
// It does not create a customer, invoice, or delivery job in the database.
func main() {
	recipient := os.Getenv("EMAIL_TEST_RECIPIENT")
	address, err := mail.ParseAddress(recipient)
	if err != nil || address.Address != recipient {
		log.Fatal("EMAIL_TEST_RECIPIENT must be one valid email address")
	}
	cfg := config.Load()
	if cfg.EmailDeliveryMode != "smtp" || cfg.EmailRendererURL == "" {
		log.Fatal("SMTP mode and EMAIL_RENDERER_URL are required")
	}
	if err := billing.ValidateInvoiceEmailConfig(cfg); err != nil {
		log.Fatal(err)
	}
	now := time.Now().UTC()
	invoice := &models.InvoiceDetailResponse{
		InvoiceListItem: models.InvoiceListItem{InvoiceNumber: "INV-UJI-TEMPLATE", CustomerName: "Penerima Uji", PlanName: "Contoh template", Amount: 0, Status: "pending", CreatedAt: now},
		Items:           []models.InvoiceItemResponse{{Description: "Uji tampilan email (bukan tagihan)", Quantity: 1, Subtotal: 0}},
	}
	settings, err := json.Marshal(map[string]string{
		"email_subject": "[Uji] Template email PayLanggan - {{invoice_number}}",
		"email_message": "Halo {{customer_name}}, ini email uji template PayLanggan. Tidak ada tagihan atau pembayaran yang dibuat.",
		"footer_note":   "Email ini hanya untuk pengujian template.",
	})
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	subject, html, _, err := billing.RenderReactInvoicePreview(ctx, cfg, invoice, "PayLanggan", recipient, settings)
	if err != nil {
		log.Fatal(err)
	}
	if err = mailer.Send(ctx, cfg, recipient, "", subject, html, "email-template-smoke-"+now.Format("20060102150405")); err != nil {
		log.Fatal(err)
	}
	log.Print("Email uji diterima oleh server SMTP")
}
