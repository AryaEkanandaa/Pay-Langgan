package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"os"
	"path/filepath"
	"pay-langgan/internal/config"
	"pay-langgan/internal/services/mailer"
	"strings"
)

func ValidateInvoiceEmailConfig(cfg *config.Config) error {
	if err := validateEmailRendererURL(cfg.EmailRendererURL); err != nil {
		return err
	}
	if cfg.EmailDeliveryMode == "disabled" || cfg.EmailDeliveryMode == "preview" {
		return nil
	}
	if cfg.EmailDeliveryMode != "smtp" {
		return fmt.Errorf("EMAIL_DELIVERY_MODE must be disabled, preview or smtp")
	}
	if _, err := mail.ParseAddress(cfg.MailFrom); err != nil || cfg.SMTPHost == "" {
		return fmt.Errorf("MAIL_FROM and SMTP_HOST must be configured")
	}
	for _, address := range []string{cfg.APIPublicURL, cfg.AppPublicURL} {
		u, err := url.Parse(address)
		if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.Hostname() == "localhost" || net.ParseIP(u.Hostname()) != nil {
			return fmt.Errorf("SMTP delivery requires public HTTPS frontend and backend URLs")
		}
	}
	return nil
}

func (s *InvoiceService) RunEmailBatch(ctx context.Context, cfg *config.Config) error {
	if err := ValidateInvoiceEmailConfig(cfg); err != nil {
		return err
	}
	if cfg.EmailDeliveryMode == "disabled" {
		return nil
	}
	if cfg.EmailDeliveryMode != "preview" && cfg.EmailDeliveryMode != "smtp" {
		return fmt.Errorf("invalid email delivery mode")
	}
	// A process interrupted during its final lease must not stay in 'sending'.
	if _, err := s.db.ExecContext(ctx, `UPDATE invoice_email_deliveries SET status='failed',last_error='delivery_failed' WHERE status IN ('queued','sending') AND next_attempt_at<=NOW() AND attempts>=5`); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE invoice_email_deliveries SET status='failed',last_error='link_expired' WHERE status='queued' AND expires_at<=NOW()`); err != nil {
		return err
	}
	for count := 0; count < 20; count++ {
		var job struct {
			ID           int             `db:"id"`
			InvoiceID    int             `db:"invoice_id"`
			BusinessID   string          `db:"business_id"`
			CustomerName string          `db:"customer_name"`
			Business     string          `db:"business_name"`
			Recipient    string          `db:"recipient"`
			Settings     json.RawMessage `db:"settings"`
			Token        string          `db:"public_token"`
			Attempts     int             `db:"attempts"`
		}
		rows, err := s.db.QueryxContext(ctx, `UPDATE invoice_email_deliveries SET status='sending',attempts=attempts+1,next_attempt_at=NOW()+INTERVAL '2 minutes' WHERE id=(SELECT id FROM invoice_email_deliveries WHERE status IN ('queued','sending') AND next_attempt_at<=NOW() AND attempts<5 ORDER BY id FOR UPDATE SKIP LOCKED LIMIT 1) RETURNING id,invoice_id,business_id,business_name,customer_name,recipient,settings,public_token,attempts`)
		if err != nil {
			return err
		}
		if !rows.Next() {
			rows.Close()
			return nil
		}
		err = rows.StructScan(&job)
		rows.Close()
		if err != nil {
			return err
		}
		inv, err := s.GetDetail(job.InvoiceID, job.BusinessID)
		subject, body, reply := "", "", ""
		if err == nil {
			inv.CustomerName = job.CustomerName
			prepared, _, prepareErr := PrepareInvoiceEmailAppearance(job.Settings, cfg)
			err = prepareErr
			if err == nil {
				// Freeze the exact appearance used by the worker for later previews.
				_, err = s.db.ExecContext(ctx, `UPDATE invoice_email_deliveries SET settings=$1 WHERE id=$2 AND status='sending' AND attempts=$3`, string(prepared), job.ID, job.Attempts)
			}
			if err == nil {
				subject, body, reply, err = renderQueuedInvoice(ctx, cfg, "invoice-issued", inv, job.Business, job.Recipient, prepared, strings.TrimRight(cfg.APIPublicURL, "/")+"/api/v1/invoice-emails/"+job.Token)
			}
		}
		status := "sent"
		if err == nil {
			_, err = s.db.ExecContext(ctx, `UPDATE invoice_email_deliveries SET rendered_subject=$1,rendered_html=$2 WHERE id=$3 AND status='sending' AND attempts=$4`, subject, body, job.ID, job.Attempts)
		}
		if err == nil && cfg.EmailDeliveryMode == "preview" {
			status = "previewed"
			err = os.MkdirAll(cfg.EmailPreviewDir, 0700)
			if err == nil {
				err = os.WriteFile(filepath.Join(cfg.EmailPreviewDir, fmt.Sprintf("invoice-%d.html", job.InvoiceID)), []byte(body), 0600)
			}
		} else if err == nil {
			err = deliverSMTP(ctx, cfg, job.Recipient, reply, subject, body, job.ID)
		}
		if err != nil {
			status = "queued"
			if job.Attempts >= 5 {
				status = "failed"
			}
			_, e := s.db.ExecContext(ctx, `UPDATE invoice_email_deliveries SET status=$1,last_error='delivery_failed',next_attempt_at=NOW()+($2 * INTERVAL '1 minute') WHERE id=$3 AND status='sending' AND attempts=$4`, status, job.Attempts*2, job.ID, job.Attempts)
			if e != nil {
				return e
			}
		} else {
			_, err = s.db.ExecContext(ctx, `UPDATE invoice_email_deliveries SET status=$1::varchar,last_error=NULL,sent_at=CASE WHEN $1::varchar='sent' THEN NOW() ELSE NULL END WHERE id=$2 AND status='sending' AND attempts=$3`, status, job.ID, job.Attempts)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
func deliverSMTP(ctx context.Context, cfg *config.Config, to, reply, subject, body string, id int) error {
	return mailer.Send(ctx, cfg, to, reply, subject, body, fmt.Sprintf("invoice-%d", id))
}
