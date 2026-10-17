package payment

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"pay-langgan/internal/utils"
	"time"
)

type WebhookLog struct {
	ID          int64     `db:"id" json:"id"`
	Provider    string    `db:"provider" json:"provider"`
	EventType   string    `db:"event_type" json:"event_type"`
	ReferenceID string    `db:"reference_id" json:"reference_id"`
	Status      string    `db:"status" json:"status"`
	HTTPStatus  int       `db:"http_status" json:"http_status"`
	ErrorCode   string    `db:"error_code" json:"error_code"`
	DurationMS  int       `db:"duration_ms" json:"duration_ms"`
	ReceivedAt  time.Time `db:"received_at" json:"received_at"`
	Processed   bool      `db:"processed" json:"processed"`
}
type WebhookLogPage struct {
	Items []WebhookLog `json:"items"`
	Total int          `json:"total"`
	Page  int          `json:"page"`
	Limit int          `json:"limit"`
}

type WebhookUnread struct {
	Count      int       `db:"count" json:"count"`
	LastSeenAt time.Time `db:"last_seen_at" json:"last_seen_at"`
	CheckedAt  time.Time `db:"checked_at" json:"checked_at"`
}

func (s *CheckoutService) NewWebhookCount(ctx context.Context, businessID string, userID int) (*WebhookUnread, error) {
	if businessID == "" || userID < 1 {
		return nil, utils.ErrUnauthorized
	}
	result := &WebhookUnread{}
	err := s.db.GetContext(ctx, result, `SELECT u.webhook_logs_last_seen_at AS last_seen_at, NOW() AS checked_at,
	(SELECT COUNT(*) FROM business_webhook_logs l WHERE l.business_id=u.business_id AND l.received_at>u.webhook_logs_last_seen_at AND l.received_at<=NOW()) AS count
	FROM users u WHERE u.id=$1 AND u.business_id=$2`, userID, businessID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, utils.ErrUnauthorized
	}
	return result, err
}

func (s *CheckoutService) MarkWebhookLogsSeen(ctx context.Context, businessID string, userID int, through time.Time) (*WebhookUnread, error) {
	if businessID == "" || userID < 1 {
		return nil, utils.ErrUnauthorized
	}
	if through.IsZero() {
		return nil, utils.ErrBadRequest
	}
	// Never move backwards (another device may have read newer logs), or into the future.
	result, err := s.db.ExecContext(ctx, `UPDATE users SET webhook_logs_last_seen_at=GREATEST(webhook_logs_last_seen_at,LEAST($1::timestamptz,NOW())) WHERE id=$2 AND business_id=$3`, through, userID, businessID)
	if err != nil {
		return nil, err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if n != 1 {
		return nil, utils.ErrUnauthorized
	}
	return s.NewWebhookCount(ctx, businessID, userID)
}

func (s *CheckoutService) ListWebhookLogs(ctx context.Context, businessID, provider, status string, page int) (*WebhookLogPage, error) {
	if businessID == "" {
		return nil, utils.ErrUnauthorized
	}
	if provider != "" && provider != "xendit" && provider != "midtrans" {
		return nil, utils.ErrBadRequest
	}
	if status != "" && status != "accepted" && status != "duplicate" && status != "failed" {
		return nil, utils.ErrBadRequest
	}
	if page < 1 || page > 100000 {
		return nil, utils.ErrBadRequest
	}
	result := &WebhookLogPage{Items: []WebhookLog{}, Page: page, Limit: 20}
	where := `business_id=$1 AND ($2='' OR provider=$2) AND ($3='' OR status=$3)`
	if err := s.db.GetContext(ctx, &result.Total, `SELECT count(*) FROM business_webhook_logs WHERE `+where, businessID, provider, status); err != nil {
		return nil, err
	}
	err := s.db.SelectContext(ctx, &result.Items, `SELECT l.id,l.provider,l.event_type,l.reference_id,l.status,l.http_status,l.error_code,l.duration_ms,l.received_at,(l.status <> 'failed' AND EXISTS(SELECT 1 FROM webhook_events e JOIN checkout_sessions cs ON cs.provider=e.provider AND (cs.reference_id=e.reference_id OR cs.provider_session_id=e.reference_id) JOIN invoices i ON i.id=cs.invoice_id JOIN subscriptions sub ON sub.id=i.subscription_id JOIN customers c ON c.id=sub.customer_id WHERE c.business_id=l.business_id AND e.provider=l.provider AND cs.reference_id=l.reference_id AND e.processed=TRUE AND (e.event_type=l.event_type OR e.event_type LIKE l.event_type||':%'))) AS processed FROM business_webhook_logs l WHERE `+where+` ORDER BY received_at DESC,id DESC LIMIT 20 OFFSET $4`, businessID, provider, status, (page-1)*20)
	return result, err
}

// Store safe delivery metadata, never raw bodies, authorization headers or keys.
func (s *CheckoutService) recordBusinessWebhook(ctx context.Context, webhookID string, payload []byte, duplicate bool, processErr error, started time.Time) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
	defer cancel()
	var owner, configuredProvider string
	if err := s.db.QueryRowContext(ctx, `SELECT business_id,provider FROM business_payment_gateways WHERE webhook_id=$1`, webhookID).Scan(&owner, &configuredProvider); err != nil {
		return
	}
	var body struct {
		Event             string `json:"event"`
		OrderID           string `json:"order_id"`
		TransactionStatus string `json:"transaction_status"`
		Data              struct {
			ReferenceID string `json:"reference_id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(payload, &body)
	event, reference, provider := body.Event, body.Data.ReferenceID, configuredProvider
	if body.OrderID != "" {
		provider = "midtrans"
		reference = body.OrderID
		event = "midtrans:" + body.TransactionStatus
	} else if body.Event != "" {
		provider = "xendit"
	}
	status, code, httpStatus := "accepted", "", 200
	if duplicate {
		status = "duplicate"
	}
	if processErr != nil {
		status = "failed"
		switch {
		case errors.Is(processErr, utils.ErrUnauthorized):
			code = "authentication_failed"
			httpStatus = 401
		case errors.Is(processErr, utils.ErrBadRequest):
			code = "invalid_event"
			httpStatus = 400
		case errors.Is(processErr, utils.ErrConflict):
			code = "state_conflict"
			httpStatus = 409
		case errors.Is(processErr, utils.ErrNotFound):
			code = "not_found"
			httpStatus = 404
		default:
			code = "processing_failed"
			httpStatus = 500
		}
	}
	if len(event) > 120 {
		event = ""
	}
	if len(reference) > 160 {
		reference = ""
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO business_webhook_logs(business_id,provider,event_type,reference_id,status,http_status,error_code,duration_ms,received_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`, owner, provider, event, reference, status, httpStatus, code, time.Since(started).Milliseconds(), started)
	if err != nil {
		slog.Error("webhook delivery log could not be stored")
	}
}
