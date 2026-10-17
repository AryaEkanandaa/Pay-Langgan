package billing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/mail"
	"pay-langgan/internal/models"
	"regexp"
	"strings"
)

type EmailAppearance struct {
	LogoWidth  int    `json:"logo_width"`
	LogoHeight int    `json:"logo_height"`
	Logo       string `json:"logo_url"`
	Color      string `json:"accent_color"`
	Address    string `json:"address"`
	Footer     string `json:"footer_note"`
	Subject    string `json:"email_subject"`
	Message    string `json:"email_message"`
	ReplyTo    string `json:"reply_to"`
}

var colorPattern = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

func rupiah(n float64) string {
	s := fmt.Sprintf("%.0f", n)
	sign := ""
	if strings.HasPrefix(s, "-") {
		sign = "-"
		s = s[1:]
	}
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "." + s[i:]
	}
	return sign + "Rp " + s
}
func RenderInvoiceEmail(invoice *models.InvoiceDetailResponse, business string, raw json.RawMessage, link string) (string, string, string, error) {
	var a EmailAppearance
	if err := json.Unmarshal(raw, &a); err != nil {
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
	if a.Subject == "" {
		a.Subject = "Invoice {{invoice_number}} - {{business_name}}"
	}
	if a.Message == "" {
		a.Message = "Halo {{customer_name}}, berikut invoice langganan Anda. Silakan periksa rincian tagihan di bawah ini."
	}
	due := "-"
	if invoice.DueDate != nil {
		due = invoice.DueDate.Format("02/01/2006")
	}
	replace := strings.NewReplacer("{{business_name}}", business, "{{customer_name}}", invoice.CustomerName, "{{invoice_number}}", invoice.InvoiceNumber, "{{total}}", rupiah(invoice.Amount), "{{due_date}}", due)
	subject := strings.ReplaceAll(strings.ReplaceAll(replace.Replace(a.Subject), "\r", " "), "\n", " ")
	a.Message = replace.Replace(a.Message)
	reply := ""
	if parsed, err := mail.ParseAddress(a.ReplyTo); err == nil && parsed.Address == a.ReplyTo {
		reply = parsed.Address
	}
	data := struct {
		Invoice    *models.InvoiceDetailResponse
		Business   string
		Appearance EmailAppearance
		Link       string
		Due        string
	}{invoice, business, a, link, due}
	// Only a validated permanent HTTPS PNG/JPG URL bypasses URL sanitization.
	t, err := template.New("invoice").Funcs(template.FuncMap{"money": rupiah, "logo": func() template.URL { return template.URL(a.Logo) }}).Parse(invoiceEmailHTML)
	if err != nil {
		return "", "", "", err
	}
	var buf bytes.Buffer
	err = t.Execute(&buf, data)
	return subject, buf.String(), reply, err
}

const invoiceEmailHTML = `<!doctype html><html lang="id"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Invoice</title></head><body style="margin:0;background:#f0f3fb;font-family:Arial,sans-serif;color:#172033"><main style="max-width:620px;margin:24px auto;background:white;padding:32px;border-radius:16px;border-top:5px solid {{.Appearance.Color}}">
{{if .Appearance.Logo}}<img src="{{logo}}" alt="Logo {{.Business}}" width="{{.Appearance.LogoWidth}}" height="{{.Appearance.LogoHeight}}" style="display:block;max-width:150px;max-height:70px;object-fit:contain;border:0">{{end}}
<h1 style="font-size:22px">{{.Business}}</h1><p style="white-space:pre-line;color:#64748b">{{.Appearance.Address}}</p>
<p style="white-space:pre-line">{{.Appearance.Message}}</p><h2>Invoice {{.Invoice.InvoiceNumber}}</h2>
<p>Pelanggan: {{.Invoice.CustomerName}}<br>Jatuh tempo: {{.Due}}<br>Status: {{if eq .Invoice.Status "paid"}}Lunas{{else if eq .Invoice.Status "pending"}}Menunggu pembayaran{{else}}{{.Invoice.Status}}{{end}}</p>
<table style="width:100%;border-collapse:collapse"><thead><tr><th style="text-align:left;padding:12px;background:#f1f5f9">Rincian</th><th style="text-align:right;padding:12px;background:#f1f5f9">Subtotal</th></tr></thead><tbody>{{range .Invoice.Items}}<tr><td style="padding:12px;border-bottom:1px solid #e2e8f0">{{.Description}} ({{.Quantity}})</td><td style="text-align:right;padding:12px;border-bottom:1px solid #e2e8f0">{{money .Subtotal}}</td></tr>{{else}}<tr><td>{{.Invoice.PlanName}}</td><td style="text-align:right">{{money .Invoice.Amount}}</td></tr>{{end}}</tbody></table>
<p style="text-align:right;font-size:20px;font-weight:bold;color:{{.Appearance.Color}}">Total {{money .Invoice.Amount}}</p>
{{if .Link}}<p><a href="{{.Link}}" style="display:inline-block;background:{{.Appearance.Color}};color:white;text-decoration:none;padding:14px 20px;border-radius:8px">Lihat invoice dan pembayaran</a></p>{{end}}
<p style="white-space:pre-line;color:#64748b">{{.Appearance.Footer}}</p><hr style="border:0;border-top:1px solid #e2e8f0"><p style="font-size:12px;color:#64748b">Dikirim melalui PayLanggan untuk {{.Business}}.</p></main></body></html>`
