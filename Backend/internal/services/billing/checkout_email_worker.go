package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"pay-langgan/internal/config"
	"pay-langgan/internal/services/mailer"
	"strings"
)

func (s *InvoiceService) RunCheckoutEmailBatch(ctx context.Context, cfg *config.Config) error {
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
	if _, err := s.db.ExecContext(ctx, `UPDATE checkout_email_deliveries SET status='failed',last_error='delivery_failed' WHERE status IN ('queued','sending') AND next_attempt_at<=NOW() AND attempts>=5`); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE checkout_email_deliveries SET status='failed',last_error='link_expired' WHERE status='queued' AND expires_at<=NOW()`); err != nil {
		return err
	}
	for count := 0; count < 20; count++ {
		var job struct {
			ID           int             `db:"id"`
			Event        string          `db:"event_type"`
			Snapshot     json.RawMessage `db:"snapshot"`
			InvoiceID    *int            `db:"invoice_id"`
			BusinessID   string          `db:"business_id"`
			CustomerName string          `db:"customer_name"`
			Business     string          `db:"business_name"`
			Recipient    string          `db:"recipient"`
			Settings     json.RawMessage `db:"settings"`
			Token        string          `db:"public_token"`
			Attempts     int             `db:"attempts"`
		}
		rows, err := s.db.QueryxContext(ctx, `UPDATE checkout_email_deliveries SET status='sending',attempts=attempts+1,next_attempt_at=NOW()+INTERVAL '2 minutes' WHERE id=(SELECT id FROM checkout_email_deliveries WHERE status IN ('queued','sending') AND next_attempt_at<=NOW() AND attempts<5 ORDER BY id FOR UPDATE SKIP LOCKED LIMIT 1) RETURNING id,event_type,snapshot,invoice_id,business_id,business_name,customer_name,recipient,settings,public_token,attempts`)
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
		if job.Event == "trial_ending" || job.Event == "invoice_due" || job.Event == "payment_failed" {
			eligible, e := false, error(nil)
			if job.Event == "payment_failed" {
				e = s.db.GetContext(ctx, &eligible, `SELECT i.status='pending' FROM invoices i WHERE i.id=$1`, job.InvoiceID)
			} else {
				eligible, e = s.reminderEligible(ctx, job.ID)
			}
			if e != nil {
				return e
			}
			if !eligible {
				if _, e = s.db.ExecContext(ctx, `UPDATE checkout_email_deliveries SET status='cancelled',last_error='reminder_not_applicable' WHERE id=$1 AND status='sending' AND attempts=$2`, job.ID, job.Attempts); e != nil {
					return e
				}
				continue
			}
		}
		prepared, _, err := PrepareInvoiceEmailAppearance(job.Settings, cfg)
		subject, body, reply := "", "", ""
		if err == nil {
			_, err = s.db.ExecContext(ctx, `UPDATE checkout_email_deliveries SET settings=$1 WHERE id=$2 AND status='sending' AND attempts=$3`, string(prepared), job.ID, job.Attempts)
		}
		if err == nil && job.Event == "trial_ending" {
			var snapshot trialEmailSnapshot
			err = s.db.QueryRowContext(ctx, `SELECT p.name,s.trial_ends_at FROM checkout_email_deliveries e JOIN subscriptions s ON s.id=e.subscription_id JOIN plans p ON p.id=s.plan_id WHERE e.id=$1`, job.ID).Scan(&snapshot.Plan, &snapshot.EndsAt)
			if err == nil {
				var raw json.RawMessage
				raw, err = json.Marshal(snapshot)
				if err == nil {
					subject, body, reply, err = renderTrialEmail(job.Business, job.CustomerName, prepared, raw, true)
				}
			}
		} else if err == nil && job.Event == "trial_started" {
			if cfg.EmailRendererURL != "" {
				var snapshot trialEmailSnapshot
				if err = json.Unmarshal(job.Snapshot, &snapshot); err == nil {
					if !cfg.AutopayEnabled {
						snapshot.AutoDebitEnabled = false
						snapshot.PaymentMethod = ""
					}
					var emailSnapshot json.RawMessage
					emailSnapshot, err = json.Marshal(snapshot)
					if err == nil {
						subject, body, reply, err = renderReactTrialStarted(ctx, cfg, job.Business, job.CustomerName, job.Recipient, prepared, emailSnapshot)
					}
				}
			} else {
				subject, body, reply, err = renderTrialConfirmation(job.Business, job.CustomerName, prepared, job.Snapshot)
			}
		} else if err == nil && job.InvoiceID != nil {
			inv, loadErr := s.GetDetail(*job.InvoiceID, job.BusinessID)
			err = loadErr
			if err == nil {
				inv.CustomerName = job.CustomerName
				link := strings.TrimRight(cfg.APIPublicURL, "/") + "/api/v1/invoice-emails/" + job.Token
				if job.Event == "invoice_due" {
					subject, body, reply, err = renderInvoiceReminder(inv, job.Business, prepared, link)
				} else if job.Event == "payment_failed" {
					var snapshot struct {
						FailureCode string `json:"failure_code"`
					}
					if err = json.Unmarshal(job.Snapshot, &snapshot); err == nil {
						subject, body, reply, err = renderQueuedInvoice(ctx, cfg, "payment-failed", inv, job.Business, job.Recipient, prepared, link, paymentFailureReason(snapshot.FailureCode))
					}
				} else {
					subject, body, reply, err = renderQueuedInvoice(ctx, cfg, "payment-success", inv, job.Business, job.Recipient, prepared, link)
				}
			}
		} else if err == nil {
			err = fmt.Errorf("missing confirmation invoice")
		}

		status := "sent"
		if err == nil {
			_, err = s.db.ExecContext(ctx, `UPDATE checkout_email_deliveries SET rendered_subject=$1,rendered_html=$2 WHERE id=$3 AND status='sending' AND attempts=$4`, subject, body, job.ID, job.Attempts)
		}
		if err == nil && cfg.EmailDeliveryMode == "preview" {
			status = "previewed"
			err = os.MkdirAll(cfg.EmailPreviewDir, 0700)
			if err == nil {
				err = os.WriteFile(filepath.Join(cfg.EmailPreviewDir, fmt.Sprintf("checkout-%d.html", job.ID)), []byte(body), 0600)
			}
		} else if err == nil {
			err = mailer.Send(ctx, cfg, job.Recipient, reply, subject, body, fmt.Sprintf("checkout-%d", job.ID))
		}
		if err != nil {
			status = "queued"
			if job.Attempts >= 5 {
				status = "failed"
			}
			_, e := s.db.ExecContext(ctx, `UPDATE checkout_email_deliveries SET status=$1,last_error='delivery_failed',next_attempt_at=NOW()+($2 * INTERVAL '1 minute') WHERE id=$3 AND status='sending' AND attempts=$4`, status, job.Attempts*2, job.ID, job.Attempts)
			if e != nil {
				return e
			}
		} else {
			_, err = s.db.ExecContext(ctx, `UPDATE checkout_email_deliveries SET status=$1::varchar,last_error=NULL,sent_at=CASE WHEN $1::varchar='sent' THEN NOW() ELSE NULL END WHERE id=$2 AND status='sending' AND attempts=$3`, status, job.ID, job.Attempts)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
