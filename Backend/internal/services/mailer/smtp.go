package mailer

import (
	"context"
	"crypto/tls"
	"fmt"
	"mime"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"net/smtp"
	"pay-langgan/internal/config"
	"strings"
	"time"
)

func Send(ctx context.Context, cfg *config.Config, to, reply, subject, body, messageID string) error {
	from, err := mail.ParseAddress(cfg.MailFrom)
	if err != nil || cfg.SMTPHost == "" {
		return fmt.Errorf("SMTP sender and host required")
	}
	target, err := mail.ParseAddress(to)
	if err != nil {
		return err
	}
	address := net.JoinHostPort(cfg.SMTPHost, cfg.SMTPPort)
	dialer := net.Dialer{Timeout: 20 * time.Second}
	var conn net.Conn
	tlsCfg := &tls.Config{ServerName: cfg.SMTPHost, MinVersion: tls.VersionTLS12}
	if cfg.SMTPImplicitTLS {
		conn, err = (&tls.Dialer{NetDialer: &dialer, Config: tlsCfg}).DialContext(ctx, "tcp", address)
	} else {
		conn, err = dialer.DialContext(ctx, "tcp", address)
	}
	if err != nil {
		return err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(30 * time.Second))
	client, err := smtp.NewClient(conn, cfg.SMTPHost)
	if err != nil {
		return err
	}
	defer client.Close()
	if !cfg.SMTPImplicitTLS {
		if ok, _ := client.Extension("STARTTLS"); !ok {
			return fmt.Errorf("SMTP requires TLS")
		}
		if err = client.StartTLS(tlsCfg); err != nil {
			return err
		}
	}
	if cfg.SMTPUsername != "" {
		if err = client.Auth(smtp.PlainAuth("", cfg.SMTPUsername, cfg.SMTPPassword, cfg.SMTPHost)); err != nil {
			return err
		}
	}
	if err = client.Mail(from.Address); err != nil {
		return err
	}
	if err = client.Rcpt(target.Address); err != nil {
		return err
	}
	writer, err := client.Data()
	if err != nil {
		return err
	}
	headers := "From: " + from.String() + "\r\nTo: " + target.String() + "\r\nSubject: " + mime.QEncoding.Encode("utf-8", subject) + "\r\nDate: " + time.Now().Format(time.RFC1123Z) + "\r\nMessage-ID: <" + messageID + "@" + strings.Split(from.Address, "@")[1] + ">\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\nContent-Transfer-Encoding: quoted-printable\r\n"
	if reply != "" {
		headers += "Reply-To: " + reply + "\r\n"
	}
	_, err = writer.Write([]byte(headers + "\r\n"))
	if err != nil {
		return err
	}
	encoded := quotedprintable.NewWriter(writer)
	if _, err = encoded.Write([]byte(body)); err != nil {
		return err
	}
	if err = encoded.Close(); err != nil {
		return err
	}
	if err = writer.Close(); err != nil {
		return err
	}
	// DATA accepted means the server owns the message; a QUIT failure is not a retry.
	_ = client.Quit()
	return nil
}
