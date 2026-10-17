package billing

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"pay-langgan/internal/config"
	"pay-langgan/internal/models"
	"strings"
	"testing"
	"time"
)

func TestInvoiceEmailEscapesBusinessAndUsesStoredAmounts(t *testing.T) {
	invoice := &models.InvoiceDetailResponse{InvoiceListItem: models.InvoiceListItem{InvoiceNumber: "INV-1", CustomerName: "<script>alert(1)</script>", Amount: 100000, Status: "pending"}, Items: []models.InvoiceItemResponse{{Description: "Professional", Quantity: 1, Subtotal: 100000}}}
	raw := json.RawMessage(`{"email_subject":"Invoice {{invoice_number}}\r\nBcc: attacker","email_message":"Halo {{customer_name}}","accent_color":"red;position:fixed","logo_url":"javascript:alert(1)","reply_to":"wrong"}`)
	subject, html, reply, err := RenderInvoiceEmail(invoice, "<b>Kelana</b>", raw, "")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(subject, "\r") || strings.Contains(subject, "\n") {
		t.Fatal("header injection")
	}
	if strings.Contains(html, "<script>") || strings.Contains(html, "<b>Kelana</b>") || strings.Contains(html, "javascript:") || strings.Contains(html, "position:fixed") {
		t.Fatal("unsafe rendering")
	}
	if !strings.Contains(html, "Rp 100.000") || !strings.Contains(html, "Menunggu pembayaran") || reply != "" {
		t.Fatal("incorrect invoice")
	}
}
func TestInvoiceLogoRasterData(t *testing.T) {
	invoice := &models.InvoiceDetailResponse{}
	_, body, _, err := RenderInvoiceEmail(invoice, "Kelana", json.RawMessage(`{"logo_url":"data:image/png;base64,aGVsbG8="}`), "")
	if err != nil || strings.Contains(body, "data:image") || strings.Contains(body, "<img") {
		t.Fatalf("embedded email logo must be omitted: %v", err)
	}
}

func TestSMTPRefusesPlaintextServer(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	done := make(chan string, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			done <- ""
			return
		}
		defer conn.Close()
		conn.SetDeadline(time.Now().Add(3 * time.Second))
		fmt.Fprint(conn, "220 localhost SMTP\r\n")
		r := bufio.NewReader(conn)
		r.ReadString('\n')
		fmt.Fprint(conn, "250 localhost\r\n")
		command, _ := r.ReadString('\n')
		done <- command
	}()
	host, port, _ := net.SplitHostPort(listener.Addr().String())
	cfg := &config.Config{MailFrom: "PayLanggan <invoice@kelana.invalid>", SMTPHost: host, SMTPPort: port}
	if err = deliverSMTP(context.Background(), cfg, "customer@kelana.invalid", "", "Invoice", "<p>Invoice</p>", 1); err == nil {
		t.Fatal("plaintext server accepted")
	}
	if command := <-done; strings.HasPrefix(command, "AUTH") || strings.HasPrefix(command, "MAIL") {
		t.Fatal("sent before TLS")
	}
}
func TestSMTPConfigRequiresSenderAndPublicURLs(t *testing.T) {
	cfg := &config.Config{EmailDeliveryMode: "preview"}
	if err := ValidateInvoiceEmailConfig(cfg); err != nil {
		t.Fatal(err)
	}
	cfg.EmailDeliveryMode = "smtp"
	if err := ValidateInvoiceEmailConfig(cfg); err == nil {
		t.Fatal("missing SMTP accepted")
	}
	cfg.MailFrom = "PayLanggan <invoice@kelana.invalid>"
	cfg.SMTPHost = "smtp.kelana.invalid"
	cfg.APIPublicURL = "http://localhost:8080"
	cfg.AppPublicURL = "https://app.kelana.invalid"
	if err := ValidateInvoiceEmailConfig(cfg); err == nil {
		t.Fatal("localhost accepted")
	}
	cfg.APIPublicURL = "https://api.kelana.invalid"
	if err := ValidateInvoiceEmailConfig(cfg); err != nil {
		t.Fatal(err)
	}
}
