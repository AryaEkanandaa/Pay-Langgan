package payment

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"pay-langgan/internal/models"
	"pay-langgan/internal/utils"
	"time"
)

// Caller must verify customer ownership before invoking this method.
// Aliases resume an existing checkout without invalidating its original token
// or creating another provider transaction (including an in-progress 3DS flow).
func (s *CheckoutService) CreatePortalLink(ctx context.Context, invoiceID int, businessID string) (*models.CheckoutLinkResponse, error) {
	tx, e := s.db.BeginTxx(ctx, nil)
	if e != nil {
		return nil, e
	}
	defer tx.Rollback()
	record, e := s.repo.FindInvoiceForCheckout(tx, invoiceID, businessID)
	if e != nil {
		return nil, e
	}
	if record == nil {
		return nil, utils.ErrNotFound
	}
	if record.InvoiceStatus != "pending" {
		return nil, utils.ErrConflict
	}
	var id int64
	var status string
	var expires time.Time
	e = tx.QueryRowContext(ctx, `SELECT id,status,expires_at FROM checkout_sessions WHERE invoice_id=$1 AND status IN ('created','creating','active') ORDER BY id DESC LIMIT 1 FOR UPDATE`, invoiceID).Scan(&id, &status, &expires)
	if e != nil && !errors.Is(e, sql.ErrNoRows) {
		return nil, e
	}
	if e == nil && (status == "creating" || (!expires.After(time.Now()) && status == "active")) {
		return nil, fmt.Errorf("%w: payment is being reconciled", utils.ErrConflict)
	}
	if e == nil && expires.After(time.Now()) {
		token, e := secureToken(32)
		if e != nil {
			return nil, e
		}
		if _, e = tx.ExecContext(ctx, `INSERT INTO portal_checkout_access(token_hash,checkout_id,expires_at) VALUES($1,$2,$3)`, hashToken(token), id, expires); e != nil {
			return nil, e
		}
		if e = tx.Commit(); e != nil {
			return nil, e
		}
		return &models.CheckoutLinkResponse{Token: token, CheckoutURL: s.publicBaseURL + "/checkout/" + token, ExpiresAt: expires}, nil
	}
	if e = tx.Commit(); e != nil {
		return nil, e
	}
	return s.CreateLink(ctx, invoiceID, businessID)
}
