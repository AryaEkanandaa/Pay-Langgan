package billing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-pdf/fpdf"
	"pay-langgan/internal/models"
)

// RenderInvoicePDF uses the same invoice and saved appearance as the public
// email link. No remote images are fetched while generating the document.
func RenderInvoicePDF(inv *models.InvoiceDetailResponse, business string, raw json.RawMessage) ([]byte, error) {
	var appearance EmailAppearance
	if err := json.Unmarshal(raw, &appearance); err != nil {
		return nil, err
	}
	if !colorPattern.MatchString(appearance.Color) {
		appearance.Color = "#3654ff"
	}
	var r, g, b int
	_, _ = fmt.Sscanf(appearance.Color, "#%02x%02x%02x", &r, &g, &b)
	// Use a compact, content-sized page so the downloaded invoice resembles
	// the public invoice card instead of leaving most of an A4 sheet empty.
	// FPDF still paginates exceptionally long invoices when they exceed this
	// single-page layout.
	pageWidth := 190.0
	pageHeight := 128.0 + float64(len(inv.Items))*10.0
	if len(inv.Items) == 0 {
		pageHeight += 10
	}
	if pageHeight > 270 {
		pageHeight = 270
	}
	pdf := fpdf.NewCustom(&fpdf.InitType{OrientationStr: "P", UnitStr: "mm", Size: fpdf.SizeType{Wd: pageWidth, Ht: pageHeight}})
	pdf.SetMargins(10, 10, 10)
	pdf.SetAutoPageBreak(true, 10)
	pdf.SetTitle("Invoice "+inv.InvoiceNumber, true)
	pdf.SetAuthor(business, true)
	translate := pdf.UnicodeTranslatorFromDescriptor("")
	text := func(value string) string { return translate(strings.ReplaceAll(value, "\t", " ")) }
	pdf.SetFooterFunc(func() {
		pdf.SetY(-8)
		pdf.SetFont("Helvetica", "", 8)
		pdf.SetTextColor(80, 91, 110)
		pdf.CellFormat(170, 5, text(fmt.Sprintf("PayLanggan | Halaman %d", pdf.PageNo())), "", 0, "R", false, 0, "")
	})
	pdf.AddPage()
	pdf.SetDrawColor(r, g, b)
	pdf.SetLineWidth(1)
	pdf.Line(10, 10, 180, 10)
	paragraph := func(value string, size float64, bold bool) {
		style := ""
		if bold {
			style = "B"
		}
		pdf.SetFont("Helvetica", style, size)
		pdf.MultiCell(170, 5, text(value), "", "L", false)
	}
	pdf.SetTextColor(23, 32, 51)
	pdf.SetXY(10, 15)
	pdf.SetFont("Helvetica", "B", 16)
	pdf.CellFormat(100, 8, text(business), "", 0, "L", false, 0, "")
	pdf.SetXY(110, 14)
	pdf.SetFont("Helvetica", "B", 18)
	pdf.SetTextColor(r, g, b)
	pdf.CellFormat(70, 8, "INVOICE", "", 1, "R", false, 0, "")
	pdf.SetXY(110, 22)
	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetTextColor(23, 32, 51)
	pdf.CellFormat(70, 6, text(inv.InvoiceNumber), "", 1, "R", false, 0, "")
	status := map[string]string{"paid": "Lunas", "pending": "Belum dibayar", "failed": "Gagal", "cancelled": "Dibatalkan"}[inv.Status]
	if status == "" {
		status = inv.Status
	}
	pdf.SetFont("Helvetica", "B", 10)
	pdf.CellFormat(70, 6, text(status), "", 1, "R", false, 0, "")
	if appearance.Address != "" {
		pdf.SetXY(10, 23)
		pdf.SetFont("Helvetica", "", 9)
		pdf.CellFormat(100, 5, text(appearance.Address), "", 1, "L", false, 0, "")
	}
	pdf.SetXY(10, 38)
	pdf.SetTextColor(90, 104, 128)
	paragraph("Ditagihkan kepada", 9, false)
	pdf.SetTextColor(23, 32, 51)
	paragraph(inv.CustomerName, 12, true)
	if inv.CustomerEmail != nil {
		paragraph(*inv.CustomerEmail, 10, false)
	}
	pdf.SetXY(110, 38)
	pdf.SetTextColor(23, 32, 51)
	rightLine := func(value string) {
		pdf.SetFont("Helvetica", "", 9)
		pdf.CellFormat(70, 5, text(value), "", 1, "R", false, 0, "")
	}
	rightLine("Terbit: " + inv.CreatedAt.Format("02/01/2006"))
	if inv.DueDate != nil {
		rightLine("Jatuh tempo: " + inv.DueDate.Format("02/01/2006"))
	}
	if inv.PaidAt != nil {
		rightLine("Dibayar: " + inv.PaidAt.Format("02/01/2006"))
	}
	if inv.Payment != nil && inv.Payment.Provider != "" {
		rightLine("Metode: " + inv.Payment.Provider)
	}
	pdf.SetXY(10, 62)
	header := func() {
		pdf.SetFont("Helvetica", "B", 10)
		pdf.SetFillColor(241, 245, 249)
		pdf.CellFormat(78, 9, "RINCIAN", "", 0, "L", true, 0, "")
		pdf.CellFormat(18, 9, "JUMLAH", "", 0, "C", true, 0, "")
		pdf.CellFormat(37, 9, "HARGA SATUAN", "", 0, "R", true, 0, "")
		pdf.CellFormat(37, 9, "SUBTOTAL", "", 1, "R", true, 0, "")
	}
	header()
	row := func(description string, quantity int, unitPrice, subtotal float64) {
		pdf.SetFont("Helvetica", "", 10)
		lines := pdf.SplitLines([]byte(text(description)), 74)
		for i, line := range lines {
			if pdf.GetY()+7 > pageHeight-12 {
				pdf.AddPage()
				header()
				pdf.SetFont("Helvetica", "", 10)
			}
			pdf.CellFormat(78, 7, string(line), "", 0, "L", false, 0, "")
			qty := ""
			if i == 0 {
				qty = fmt.Sprint(quantity)
			}
			pdf.CellFormat(18, 7, qty, "", 0, "C", false, 0, "")
			unit := ""
			amount := ""
			if i == 0 {
				unit = rupiah(unitPrice)
				amount = rupiah(subtotal)
			}
			pdf.CellFormat(37, 7, unit, "", 0, "R", false, 0, "")
			pdf.CellFormat(37, 7, amount, "", 1, "R", false, 0, "")
		}
		pdf.SetDrawColor(226, 232, 240)
		pdf.SetLineWidth(0.2)
		pdf.Line(10, pdf.GetY()+1, 180, pdf.GetY()+1)
		pdf.Ln(3)
	}
	if len(inv.Items) == 0 {
		row(inv.PlanName, 1, inv.Amount, inv.Amount)
	}
	for _, item := range inv.Items {
		row(item.Description, item.Quantity, item.UnitPrice, item.Subtotal)
	}
	if pdf.GetY()+20 > pageHeight-12 {
		pdf.AddPage()
	}
	pdf.Ln(5)
	pdf.SetFont("Helvetica", "B", 15)
	pdf.SetTextColor(r, g, b)
	pdf.CellFormat(170, 10, "Total "+rupiah(inv.Amount), "", 1, "R", false, 0, "")
	pdf.Ln(8)
	pdf.SetTextColor(80, 91, 110)
	if appearance.Footer != "" {
		paragraph(appearance.Footer, 10, false)
		pdf.Ln(4)
	}
	paragraph("Dikirim melalui PayLanggan untuk "+business+".", 9, false)
	var output bytes.Buffer
	if err := pdf.Output(&output); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}
