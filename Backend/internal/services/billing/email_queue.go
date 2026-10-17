package billing

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
	"net/mail"
	"pay-langgan/internal/utils"
	"time"
)

type InvoiceEmailStatus struct {
	Status    string     `db:"status" json:"status"`
	Recipient string     `db:"recipient" json:"recipient"`
	Attempts  int        `db:"attempts" json:"attempts"`
	LastError *string    `db:"last_error" json:"last_error"`
	SentAt    *time.Time `db:"sent_at" json:"sent_at"`
}

// Automatic jobs are created in the invoice transaction, after its line items.
// Existing invoices are never bulk-mailed when a business enables this option.
func queueInvoiceEmail(ctx context.Context, tx *sqlx.Tx, invoiceID int, manual bool) error {
	var owner struct {
		BusinessID   string          `db:"business_id"`
		BusinessName string          `db:"business_name"`
		CustomerName string          `db:"customer_name"`
		Recipient    string          `db:"recipient"`
		Settings     json.RawMessage `db:"settings"`
	}
	err := tx.GetContext(ctx, &owner, `SELECT c.business_id,b.name business_name,c.name customer_name,COALESCE(c.email,'') recipient,COALESCE(b.meta->'invoice','{}'::jsonb) settings FROM invoices i JOIN subscriptions s ON s.id=i.subscription_id JOIN customers c ON c.id=s.customer_id JOIN businesses b ON b.id=c.business_id WHERE i.id=$1`, invoiceID)
	if err != nil {
		return err
	}
	var options struct {
		AutoSend bool `json:"auto_send"`
	}
	_ = json.Unmarshal(owner.Settings, &options)
	if !manual && !options.AutoSend {
		return nil
	}
	address, err := mail.ParseAddress(owner.Recipient)
	if err != nil || address.Address != owner.Recipient {
		if manual {
			return fmt.Errorf("%w: customer email is missing or invalid", utils.ErrBadRequest)
		}
		return nil
	}
	bytes := make([]byte, 32)
	if _, err = rand.Read(bytes); err != nil {
		return err
	}
	token := base64.RawURLEncoding.EncodeToString(bytes)
	_, err = tx.ExecContext(ctx, `INSERT INTO invoice_email_deliveries(invoice_id,business_id,business_name,customer_name,recipient,settings,public_token) VALUES($1,$2,$3,$4,$5,$6,$7) ON CONFLICT(invoice_id) DO UPDATE SET status='queued',attempts=0,last_error=NULL,rendered_subject=NULL,rendered_html=NULL,next_attempt_at=NOW(),business_name=EXCLUDED.business_name,customer_name=EXCLUDED.customer_name,recipient=EXCLUDED.recipient,settings=EXCLUDED.settings,public_token=EXCLUDED.public_token,expires_at=NOW()+INTERVAL '90 days' WHERE invoice_email_deliveries.status IN ('failed','previewed')`, invoiceID, owner.BusinessID, owner.BusinessName, owner.CustomerName, owner.Recipient, string(owner.Settings), token)
	return err
}

func (s *InvoiceService) QueueEmail(ctx context.Context, id int, businessID string) (*InvoiceEmailStatus, error) {
	if _, err := s.GetDetail(id, businessID); err != nil {
		return nil, err
	}
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	// Serialize requests for the same invoice and protect against duplicate clicks.
	if _, err = tx.ExecContext(ctx, `SELECT id FROM invoices WHERE id=$1 FOR UPDATE`, id); err != nil {
		return nil, err
	}
	if err = queueInvoiceEmail(ctx, tx, id, true); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return s.EmailStatus(ctx, id, businessID)
}
func (s *InvoiceService) EmailStatus(ctx context.Context, id int, businessID string) (*InvoiceEmailStatus, error) {
	if _, err := s.GetDetail(id, businessID); err != nil {
		return nil, err
	}
	var result InvoiceEmailStatus
	rows, err := s.db.QueryxContext(ctx, `SELECT status,recipient,attempts,last_error,sent_at FROM invoice_email_deliveries WHERE invoice_id=$1 AND business_id=$2`, id, businessID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, nil
	}
	if err = rows.StructScan(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *InvoiceService) PublicEmailInvoice(ctx context.Context, token string) (int, string, json.RawMessage, error) {
	if len(token) != 43 {
		return 0, "", nil, utils.ErrNotFound
	}
	var id int
	var businessID string
	var settings json.RawMessage
	err := s.db.QueryRowContext(ctx, `SELECT invoice_id,business_id,settings FROM invoice_email_deliveries WHERE public_token=$1 AND expires_at>NOW()
 UNION ALL SELECT invoice_id,business_id,settings FROM checkout_email_deliveries WHERE public_token=$1 AND expires_at>NOW() AND event_type IN ('payment_succeeded','invoice_due') LIMIT 1`, token).Scan(&id, &businessID, &settings)
	if err != nil {
		return 0, "", nil, utils.ErrNotFound
	}
	return id, businessID, settings, nil
}

func (s *InvoiceService) EmailSettings(ctx context.Context, businessID string, name *string, raw *json.RawMessage) error {
	return s.db.QueryRowContext(ctx, `SELECT name,COALESCE(meta->'invoice','{}'::jsonb) FROM businesses WHERE id=$1`, businessID).Scan(name, raw)
}

// Preview the queued/sent snapshot, not a business appearance edited afterwards.
func (s *InvoiceService) EmailPreviewSettings(ctx context.Context, id int, businessID string, name, customer *string, raw *json.RawMessage) error {
	err := s.db.QueryRowContext(ctx, `SELECT business_name,customer_name,settings FROM invoice_email_deliveries WHERE invoice_id=$1 AND business_id=$2`, id, businessID).Scan(name, customer, raw)
	if err == sql.ErrNoRows {
		return s.EmailSettings(ctx, businessID, name, raw)
	}
	return err
}
