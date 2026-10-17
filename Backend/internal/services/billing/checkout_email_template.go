package billing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/mail"
	"pay-langgan/internal/models"
	"strings"
	"time"
)

func renderPaymentConfirmation(inv *models.InvoiceDetailResponse, business string, raw json.RawMessage, link string) (string, string, string, error) {
	if inv.Status != "paid" {
		return "", "", "", fmt.Errorf("receipt requires a paid invoice")
	}
	var appearance map[string]interface{}
	if err := json.Unmarshal(raw, &appearance); err != nil {
		return "", "", "", err
	}
	appearance["email_subject"] = confirmationSubject("payment_succeeded", business, inv.InvoiceNumber)
	message := "Halo {{customer_name}}, pembayaran untuk {{business_name}} telah berhasil. Berikut invoice lunas Anda."
	if inv.Amount == 0 {
		message = "Halo {{customer_name}}, checkout Anda selesai. Total tagihan Rp 0, sehingga tidak ada dana yang ditagihkan ke kartu Anda. Berikut invoice lunas Anda."
	}
	if inv.PaidAt != nil {
		message += "\nTanggal pembayaran: " + inv.PaidAt.UTC().Format("02/01/2006 15:04") + " UTC."
	}
	if inv.PeriodStart != nil && inv.PeriodEnd != nil {
		message += "\nPeriode tagihan: " + inv.PeriodStart.Format("02/01/2006") + " – " + inv.PeriodEnd.Format("02/01/2006") + "."
	}
	appearance["email_message"] = message
	prepared, err := json.Marshal(appearance)
	if err != nil {
		return "", "", "", err
	}
	return RenderInvoiceEmail(inv, business, prepared, link)
}

func renderInvoiceReminder(inv *models.InvoiceDetailResponse, business string, raw json.RawMessage, link string) (string, string, string, error) {
	if inv.Status != "pending" {
		return "", "", "", fmt.Errorf("reminder requires an unpaid invoice")
	}
	var appearance map[string]interface{}
	if err := json.Unmarshal(raw, &appearance); err != nil {
		return "", "", "", err
	}
	appearance["email_subject"] = "Pengingat invoice {{invoice_number}} - {{business_name}}"
	appearance["email_message"] = "Halo {{customer_name}}, invoice {{invoice_number}} sebesar {{total}} akan jatuh tempo pada {{due_date}}. Silakan lihat rincian dan lanjutkan pembayaran melalui tautan di bawah. Jika Anda baru menyelesaikan pembayaran, tunggu konfirmasi pembayaran dari PayLanggan."
	prepared, err := json.Marshal(appearance)
	if err != nil {
		return "", "", "", err
	}
	return RenderInvoiceEmail(inv, business, prepared, link)
}

func renderTrialConfirmation(business, customer string, raw, snapshot json.RawMessage) (string, string, string, error) {
	return renderTrialEmail(business, customer, raw, snapshot, false)
}
func renderTrialEmail(business, customer string, raw, snapshot json.RawMessage, reminder bool) (string, string, string, error) {
	var a EmailAppearance
	var trial trialEmailSnapshot
	if err := json.Unmarshal(raw, &a); err != nil {
		return "", "", "", err
	}
	if err := json.Unmarshal(snapshot, &trial); err != nil {
		return "", "", "", err
	}
	if !colorPattern.MatchString(a.Color) {
		a.Color = "#3654ff"
	}
	if !ValidEmailLogoURL(a.Logo) {
		a.Logo = ""
	}
	if a.LogoWidth < 1 || a.LogoWidth > 150 {
		a.LogoWidth = 150
	}
	if a.LogoHeight < 1 || a.LogoHeight > 70 {
		a.LogoHeight = 70
	}
	reply := ""
	if address, err := mail.ParseAddress(a.ReplyTo); err == nil && address.Address == a.ReplyTo {
		reply = a.ReplyTo
	}
	data := struct {
		Reminder                        bool
		Business, Customer, Ends, Cycle string
		Appearance                      EmailAppearance
		Trial                           trialEmailSnapshot
	}{reminder, business, customer, trial.EndsAt.In(time.UTC).Format("02/01/2006 15:04 UTC"), "bulan", a, trial}
	if trial.Cycle == "yearly" {
		data.Cycle = "tahun"
	}
	t, err := template.New("trial").Funcs(template.FuncMap{"money": rupiah, "logo": func() template.URL { return template.URL(a.Logo) }}).Parse(trialConfirmationHTML)
	if err != nil {
		return "", "", "", err
	}
	var buf bytes.Buffer
	err = t.Execute(&buf, data)
	subject := strings.NewReplacer("\r", " ", "\n", " ").Replace(confirmationSubject("trial_started", business, ""))
	if reminder {
		subject = strings.NewReplacer("\r", " ", "\n", " ").Replace("Trial segera berakhir - " + business)
	}
	return subject, buf.String(), reply, err
}

const trialConfirmationHTML = `<!doctype html><html lang="id"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Konfirmasi trial</title></head><body style="margin:0;background:#f0f3fb;font-family:Arial,sans-serif;color:#172033"><main style="max-width:620px;margin:24px auto;background:white;padding:32px;border-radius:16px;border-top:5px solid {{.Appearance.Color}}">
{{if .Appearance.Logo}}<img src="{{logo}}" alt="Logo {{.Business}}" width="{{.Appearance.LogoWidth}}" height="{{.Appearance.LogoHeight}}" style="display:block;max-width:150px;max-height:70px;object-fit:contain;border:0">{{end}}
<h1 style="font-size:22px">{{.Business}}</h1><p style="white-space:pre-line;color:#475569">{{.Appearance.Address}}</p>
{{if .Reminder}}<h2>Trial segera berakhir</h2><p>Halo {{.Customer}}, masa trial paket <strong>{{.Trial.Plan}}</strong> akan segera berakhir. Hubungi {{.Business}} untuk melanjutkan langganan.</p>{{else}}<h2>Trial berhasil dimulai</h2><p>Halo {{.Customer}}, trial paket <strong>{{.Trial.Plan}}</strong> sudah dimulai.</p>{{end}}
<p>Trial berakhir: <strong>{{.Ends}}</strong></p><p>Tagihan sekarang: <strong>Rp 0</strong>. Tidak ada biaya paket yang ditagihkan hari ini.</p>
{{if not .Reminder}}<table style="width:100%;border-collapse:collapse"><thead><tr><th style="text-align:left;padding:12px;background:#f1f5f9">Rincian setelah trial</th><th style="text-align:right;padding:12px;background:#f1f5f9">Subtotal</th></tr></thead><tbody>{{range .Trial.Items}}<tr><td style="padding:12px;border-bottom:1px solid #e2e8f0">{{.Name}} ({{.Quantity}})</td><td style="text-align:right;padding:12px;border-bottom:1px solid #e2e8f0">{{money .Subtotal}}</td></tr>{{end}}</tbody></table>
<p style="font-size:18px;font-weight:bold;color:{{.Appearance.Color}}">Estimasi setelah trial: {{money .Trial.Amount}} / {{.Cycle}}</p>{{end}}
{{if .Trial.AutoDebitEnabled}}<p>Setelah trial berakhir, invoice akan dibuat dan kartu yang terdaftar di {{if eq .Trial.PaymentMethod "midtrans"}}Midtrans{{else}}Xendit{{end}} akan didebit otomatis sebesar {{money .Trial.Amount}} per {{.Cycle}} sesuai persetujuan Anda. Jika debit gagal, PayLanggan akan mengirimkan invoice agar Anda dapat melanjutkan pembayaran.</p>{{else}}<p>Invoice akan dibuat setelah trial berakhir dan perlu dibayar secara manual. Harga dan kupon mengikuti ketentuan yang berlaku saat penagihan.</p>{{end}}<p>Hubungi {{.Business}} untuk informasi kelanjutan langganan dan pembayaran.</p>
<p style="white-space:pre-line;color:#475569">{{.Appearance.Footer}}</p><hr style="border:0;border-top:1px solid #e2e8f0"><p style="font-size:12px;color:#475569">Dikirim melalui PayLanggan untuk {{.Business}}.</p></main></body></html>`
