package billing

import (
	"context"
	"database/sql"
	"fmt"
	"pay-langgan/internal/utils"
	"strings"
	"time"
)

type CustomerEmailLog struct {
	ID           int        `db:"id" json:"id"`
	Source       string     `db:"source" json:"source"`
	Kind         string     `db:"kind" json:"kind"`
	InvoiceID    *int       `db:"invoice_id" json:"invoice_id"`
	CustomerName string     `db:"customer_name" json:"customer_name"`
	Recipient    string     `db:"recipient" json:"recipient"`
	Status       string     `db:"status" json:"status"`
	Attempts     int        `db:"attempts" json:"attempts"`
	LastError    *string    `db:"last_error" json:"last_error,omitempty"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	SentAt       *time.Time `db:"sent_at" json:"sent_at,omitempty"`
}

const emailHistoryUnion = `SELECT id,'invoice' source,'invoice_issued' kind,invoice_id,customer_name,recipient,status,attempts,last_error,created_at,sent_at FROM invoice_email_deliveries WHERE business_id=$1 UNION ALL SELECT id,'checkout' source,event_type kind,invoice_id,customer_name,recipient,status,attempts,last_error,created_at,sent_at FROM checkout_email_deliveries WHERE business_id=$1`

func (s *InvoiceService) CustomerEmailHistory(ctx context.Context, business, status, kind, search string, page int) ([]CustomerEmailLog, int, error) {
	if page < 1 {
		page = 1
	}
	if page > 100000 {
		return nil, 0, utils.ErrBadRequest
	}
	if status != "" && !strings.Contains("|queued|sending|sent|previewed|failed|cancelled|", "|"+status+"|") {
		return nil, 0, utils.ErrBadRequest
	}
	if kind != "" && !strings.Contains("|invoice_issued|payment_succeeded|trial_started|trial_ending|invoice_due|", "|"+kind+"|") {
		return nil, 0, utils.ErrBadRequest
	}
	filter := ` FROM (` + emailHistoryUnion + `) e WHERE ($2='' OR status=$2) AND ($3='' OR kind=$3) AND ($4='' OR recipient ILIKE '%'||$4||'%' OR customer_name ILIKE '%'||$4||'%')`
	args := []interface{}{business, status, kind, strings.TrimSpace(search)}
	var total int
	if err := s.db.GetContext(ctx, &total, `SELECT COUNT(*)`+filter, args...); err != nil {
		return nil, 0, err
	}
	result := []CustomerEmailLog{}
	err := s.db.SelectContext(ctx, &result, `SELECT *`+filter+` ORDER BY created_at DESC,source,id DESC LIMIT 20 OFFSET $5`, append(args, (page-1)*20)...)
	return result, total, err
}
func emailTable(source string) (string, error) {
	if source == "invoice" {
		return "invoice_email_deliveries", nil
	}
	if source == "checkout" {
		return "checkout_email_deliveries", nil
	}
	return "", utils.ErrBadRequest
}

func (s *InvoiceService) RetryCustomerEmail(ctx context.Context, business, source string, id int) error {
	table, err := emailTable(source)
	if err != nil {
		return err
	}
	result, err := s.db.ExecContext(ctx, `UPDATE `+table+` SET status='queued',attempts=0,last_error=NULL,next_attempt_at=NOW(),expires_at=NOW()+INTERVAL '90 days',rendered_subject=NULL,rendered_html=NULL WHERE id=$1 AND business_id=$2 AND status='failed'`, id, business)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return fmt.Errorf("%w: hanya email gagal yang dapat dicoba ulang", utils.ErrConflict)
	}
	return nil
}

func (s *InvoiceService) CustomerEmailPreview(ctx context.Context, business, source string, id int) (string, string, error) {
	table, err := emailTable(source)
	if err != nil {
		return "", "", err
	}
	var subject, body sql.NullString
	err = s.db.QueryRowContext(ctx, `SELECT rendered_subject,rendered_html FROM `+table+` WHERE id=$1 AND business_id=$2`, id, business).Scan(&subject, &body)
	if err == sql.ErrNoRows {
		return "", "", utils.ErrNotFound
	}
	if err != nil {
		return "", "", err
	}
	if !body.Valid {
		return "", "", fmt.Errorf("%w: preview tersedia setelah email diproses", utils.ErrConflict)
	}
	return subject.String, body.String, nil
}
