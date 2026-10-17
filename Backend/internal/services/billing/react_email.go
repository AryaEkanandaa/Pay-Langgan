package billing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"net/url"
	"pay-langgan/internal/config"
	"pay-langgan/internal/models"
	"strings"
	"time"
)

// The React renderer is a private loopback service. Rendering failures are
// returned to the existing delivery queue, which retains its retry policy.
func validateEmailRendererURL(value string) error {
	if value == "" {
		return nil
	}
	u, err := url.Parse(value)
	if err != nil || u.Scheme != "http" || u.Hostname() != "127.0.0.1" || u.Port() == "" || u.Path != "" || u.RawQuery != "" || u.Fragment != "" || u.User != nil {
		return fmt.Errorf("EMAIL_RENDERER_URL must be an http://127.0.0.1:port address")
	}
	return nil
}

type reactInvoiceItem struct {
	Name      string  `json:"name"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unitPrice"`
	Subtotal  float64 `json:"subtotal"`
}

func emailDate(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.Format(time.RFC3339)
}

func paymentFailureReason(code string) string {
	switch strings.ToLower(code) {
	case "insufficient_funds", "insufficient_balance":
		return "Saldo atau limit kartu tidak mencukupi. Gunakan kartu lain atau hubungi bank penerbit, lalu coba lagi."
	case "expired_card", "card_expired":
		return "Masa berlaku kartu telah habis. Perbarui metode pembayaran dengan kartu yang masih berlaku."
	case "card_declined", "declined", "do_not_honor":
		return "Bank penerbit menolak pembayaran kartu. Hubungi bank penerbit atau gunakan kartu lain, lalu coba lagi."
	default:
		return "Pembayaran kartu belum berhasil. Periksa kartu Anda atau hubungi bank penerbit, lalu coba lagi."
	}
}

func invoiceReactProps(inv *models.InvoiceDetailResponse, business, recipient string, a EmailAppearance, link string) map[string]interface{} {
	items := make([]reactInvoiceItem, 0, len(inv.Items))
	subtotal, discount := float64(0), float64(0)
	for _, item := range inv.Items {
		quantity := item.Quantity
		if quantity < 1 {
			quantity = 1
		}
		unit := item.UnitPrice
		if unit == 0 {
			unit = item.Subtotal / float64(quantity)
		}
		items = append(items, reactInvoiceItem{item.Description, quantity, unit, item.Subtotal})
		if item.Subtotal < 0 {
			discount -= item.Subtotal
		} else {
			subtotal += item.Subtotal
		}
	}
	if len(items) == 0 {
		items = append(items, reactInvoiceItem{inv.PlanName, 1, inv.Amount, inv.Amount})
		subtotal = inv.Amount
	}
	status := inv.Status
	switch status {
	case "paid", "pending", "overdue", "failed", "cancelled", "refunded":
	default:
		status = "pending"
	}
	invoice := map[string]interface{}{
		"number": inv.InvoiceNumber, "status": status,
		"issuedAt": inv.CreatedAt.Format(time.RFC3339), "dueDate": emailDate(inv.DueDate),
		"paidAt": emailDate(inv.PaidAt), "periodStart": emailDate(inv.PeriodStart),
		"periodEnd": emailDate(inv.PeriodEnd), "items": items,
		"subtotal": subtotal, "discountTotal": discount,
		"total": inv.Amount, "currency": "IDR", "url": link,
	}
	if inv.Payment != nil {
		invoice["paymentMethod"] = inv.Payment.Provider
	}
	return map[string]interface{}{
		"brand":      map[string]string{"name": business, "logoUrl": a.Logo, "brandColor": a.Color, "address": a.Address},
		"customer":   map[string]string{"name": inv.CustomerName, "email": recipient},
		"invoice":    invoice,
		"footerNote": a.Footer,
	}
}

func renderReactEmail(ctx context.Context, cfg *config.Config, kind string, inv *models.InvoiceDetailResponse, business, recipient string, raw json.RawMessage, link, failureReason string, publicActions map[string]string) (string, string, string, error) {
	if err := validateEmailRendererURL(cfg.EmailRendererURL); err != nil {
		return "", "", "", err
	}
	var appearance EmailAppearance
	if err := json.Unmarshal(raw, &appearance); err != nil {
		return "", "", "", err
	}
	if !colorPattern.MatchString(appearance.Color) {
		appearance.Color = "#3B4EFF"
	}
	if !ValidEmailLogoURL(appearance.Logo) {
		appearance.Logo = ""
	}
	props := invoiceReactProps(inv, business, recipient, appearance, link)
	if publicActions != nil {
		props["publicActions"] = publicActions
	}
	if kind == "invoice-issued" {
		due := "-"
		if inv.DueDate != nil {
			due = inv.DueDate.Format("02/01/2006")
		}
		replacer := strings.NewReplacer("{{business_name}}", business, "{{customer_name}}", inv.CustomerName, "{{invoice_number}}", inv.InvoiceNumber, "{{total}}", rupiah(inv.Amount), "{{due_date}}", due)
		if appearance.Subject != "" {
			props["subjectOverride"] = strings.NewReplacer("\r", " ", "\n", " ").Replace(replacer.Replace(appearance.Subject))
		}
		if appearance.Message != "" {
			props["messageOverride"] = replacer.Replace(appearance.Message)
		}
	} else if kind == "payment-failed" {
		props["failureReason"] = failureReason
		props["actionUrl"] = link
	}
	payload, err := json.Marshal(map[string]interface{}{"type": kind, "props": props})
	if err != nil {
		return "", "", "", err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.EmailRendererURL+"/render", bytes.NewReader(payload))
	if err != nil {
		return "", "", "", err
	}
	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 8 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return "", "", "", fmt.Errorf("email renderer unavailable: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", "", "", fmt.Errorf("email renderer returned HTTP %d", response.StatusCode)
	}
	var rendered struct {
		Subject string `json:"subject"`
		HTML    string `json:"html"`
	}
	if err = json.NewDecoder(io.LimitReader(response.Body, 2<<20)).Decode(&rendered); err != nil {
		return "", "", "", err
	}
	if rendered.Subject == "" || rendered.HTML == "" {
		return "", "", "", fmt.Errorf("email renderer returned empty content")
	}
	reply := ""
	if address, err := mail.ParseAddress(appearance.ReplyTo); err == nil && address.Address == appearance.ReplyTo {
		reply = address.Address
	}
	return rendered.Subject, rendered.HTML, reply, nil
}

func renderQueuedInvoice(ctx context.Context, cfg *config.Config, kind string, inv *models.InvoiceDetailResponse, business, recipient string, raw json.RawMessage, link string, failureReason ...string) (string, string, string, error) {
	if cfg.EmailRendererURL != "" {
		reason := "Pembayaran kartu belum berhasil. Periksa kartu Anda atau hubungi bank penerbit, lalu coba lagi."
		if len(failureReason) > 0 && failureReason[0] != "" {
			reason = failureReason[0]
		}
		return renderReactEmail(ctx, cfg, kind, inv, business, recipient, raw, link, reason, nil)
	}
	if kind == "payment-success" {
		return renderPaymentConfirmation(inv, business, raw, link)
	}
	if kind == "payment-failed" {
		var appearance map[string]interface{}
		if err := json.Unmarshal(raw, &appearance); err != nil {
			return "", "", "", err
		}
		appearance["email_subject"] = "Pembayaran gagal {{total}} - {{invoice_number}} - {{business_name}}"
		appearance["email_message"] = "Halo {{customer_name}}, pembayaran invoice {{invoice_number}} belum berhasil. Periksa kartu Anda atau hubungi bank penerbit, lalu coba lagi melalui tautan invoice."
		prepared, err := json.Marshal(appearance)
		if err != nil {
			return "", "", "", err
		}
		return RenderInvoiceEmail(inv, business, prepared, link)
	}
	return RenderInvoiceEmail(inv, business, raw, link)
}

func renderReactTrialStarted(ctx context.Context, cfg *config.Config, business, customer, recipient string, appearanceRaw, snapshotRaw json.RawMessage) (string, string, string, error) {
	if err := validateEmailRendererURL(cfg.EmailRendererURL); err != nil {
		return "", "", "", err
	}
	var appearance EmailAppearance
	if err := json.Unmarshal(appearanceRaw, &appearance); err != nil {
		return "", "", "", err
	}
	if !colorPattern.MatchString(appearance.Color) {
		appearance.Color = "#3B4EFF"
	}
	if !ValidEmailLogoURL(appearance.Logo) {
		appearance.Logo = ""
	}
	var snapshot trialEmailSnapshot
	if err := json.Unmarshal(snapshotRaw, &snapshot); err != nil {
		return "", "", "", err
	}
	items := make([]map[string]interface{}, 0, len(snapshot.Items))
	for _, item := range snapshot.Items {
		items = append(items, map[string]interface{}{"name": item.Name, "quantity": item.Quantity, "subtotal": item.Subtotal})
	}
	props := map[string]interface{}{
		"brand":    map[string]string{"name": business, "logoUrl": appearance.Logo, "brandColor": appearance.Color, "address": appearance.Address},
		"customer": map[string]string{"name": customer, "email": recipient},
		"trial": map[string]interface{}{"plan": snapshot.Plan, "endsAt": snapshot.EndsAt.Format(time.RFC3339), "billingCycle": snapshot.Cycle,
			"amount": snapshot.Amount, "items": items, "autoDebitEnabled": snapshot.AutoDebitEnabled, "paymentMethod": snapshot.PaymentMethod},
		"footerNote": appearance.Footer,
	}
	payload, err := json.Marshal(map[string]interface{}{"type": "trial-started", "props": props})
	if err != nil {
		return "", "", "", err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.EmailRendererURL+"/render", bytes.NewReader(payload))
	if err != nil {
		return "", "", "", err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := (&http.Client{Timeout: 8 * time.Second}).Do(request)
	if err != nil {
		return "", "", "", fmt.Errorf("email renderer unavailable: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", "", "", fmt.Errorf("email renderer returned HTTP %d", response.StatusCode)
	}
	var rendered struct {
		Subject string `json:"subject"`
		HTML    string `json:"html"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 2<<20)).Decode(&rendered); err != nil {
		return "", "", "", err
	}
	if rendered.Subject == "" || rendered.HTML == "" {
		return "", "", "", fmt.Errorf("email renderer returned empty content")
	}
	reply := ""
	if address, err := mail.ParseAddress(appearance.ReplyTo); err == nil && address.Address == appearance.ReplyTo {
		reply = address.Address
	}
	return rendered.Subject, rendered.HTML, reply, nil
}

func RenderReactInvoicePreview(ctx context.Context, cfg *config.Config, inv *models.InvoiceDetailResponse, business, recipient string, raw json.RawMessage) (string, string, string, error) {
	return renderReactEmail(ctx, cfg, "invoice-issued", inv, business, recipient, raw, "", "", nil)
}

func RenderReactPublicInvoice(ctx context.Context, cfg *config.Config, inv *models.InvoiceDetailResponse, business, recipient string, raw json.RawMessage, checkoutPath, downloadPath string) (string, string, string, error) {
	return renderReactEmail(ctx, cfg, "invoice-issued", inv, business, recipient, raw, "", "", map[string]string{
		"downloadPath": downloadPath,
		"checkoutPath": checkoutPath,
	})
}
