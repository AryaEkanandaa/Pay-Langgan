package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"pay-langgan/internal/config"
	"pay-langgan/internal/models"
)

type AppearancePreview struct {
	Subject     string `json:"subject"`
	HTML        string `json:"html"`
	LogoWarning string `json:"logo_warning"`
}

func appearanceSampleInvoice(now time.Time) *models.InvoiceDetailResponse {
	end := now.AddDate(0, 1, 0)
	return &models.InvoiceDetailResponse{
		InvoiceListItem: models.InvoiceListItem{InvoiceNumber: "INV-CONTOH-0001", CustomerName: "Pelanggan Contoh", PlanName: "Paket Contoh", Amount: 218500, Status: "pending", CreatedAt: now, DueDate: &now, PeriodStart: &now, PeriodEnd: &end},
		Items: []models.InvoiceItemResponse{
			{Description: "Paket Contoh", Quantity: 1, Subtotal: 200000},
			{Description: "Add-on Contoh", Quantity: 1, Subtotal: 30000},
			{Description: "Coupon CONTOH", Quantity: 1, Subtotal: -11500},
		},
	}
}

// Render the draft using exactly the same asset preparation and templates as
// the email worker. Sample data never creates an invoice or queues an email.
func PreviewInvoiceAppearance(business string, raw json.RawMessage, kind string, cfg *config.Config, now time.Time) (*AppearancePreview, error) {
	if kind != "invoice_issued" && kind != "payment_succeeded" {
		return nil, fmt.Errorf("invalid preview kind")
	}
	prepared, warning, err := PrepareInvoiceEmailAppearance(raw, cfg)
	if err != nil {
		return nil, err
	}
	inv := appearanceSampleInvoice(now)
	link := "#invoice-contoh"
	if cfg.EmailRendererURL != "" {
		link = "https://paylanggan.example/invoice-contoh"
	}
	var subject, body string
	if kind == "payment_succeeded" {
		inv.Status = "paid"
		inv.PaidAt = &now
		subject, body, _, err = renderQueuedInvoice(context.Background(), cfg, "payment-success", inv, business, "pelanggan@contoh.example", prepared, link)
	} else {
		subject, body, _, err = renderQueuedInvoice(context.Background(), cfg, "invoice-issued", inv, business, "pelanggan@contoh.example", prepared, link)
	}
	if err != nil {
		return nil, err
	}
	return &AppearancePreview{Subject: subject, HTML: body, LogoWarning: warning}, nil
}
