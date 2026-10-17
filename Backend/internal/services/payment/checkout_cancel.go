package payment

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"pay-langgan/internal/models"
	"pay-langgan/internal/utils"
	provider "pay-langgan/pkg/payment"
)

type xenditSessionCanceller interface {
	CancelCardSession(context.Context, string, string) (*provider.XenditSessionResponse, error)
}

// Cancel closes only an unfinished checkout. A payment being created or
// verified cannot be discarded because it might still settle at the gateway.
func (s *CheckoutService) Cancel(ctx context.Context, token string) (*models.PublicCheckoutResponse, error) {
	if strings.HasPrefix(token, "trial_") {
		if err := s.cancelTrialCard(ctx, token); err != nil {
			return nil, err
		}
		return s.Get(ctx, token)
	}
	if err := s.cancelInvoiceCheckout(ctx, token); err != nil {
		return nil, err
	}
	return s.Get(ctx, token)
}

func (s *CheckoutService) cancelXenditSession(ctx context.Context, businessID, sessionID, reference, purpose string, amount int64) error {
	account, err := s.xenditForBusiness(ctx, businessID)
	if err != nil {
		return err
	}
	session, err := account.GetCardSession(ctx, sessionID, reference)
	if err != nil {
		return err
	}
	if session.PaymentSessionID != sessionID || session.ReferenceID != reference || session.SessionType != purpose || session.Currency != "IDR" || session.Amount != amount {
		return fmt.Errorf("%w: sesi gateway tidak sesuai", utils.ErrConflict)
	}
	switch session.Status {
	case "CANCELED", "EXPIRED":
		return nil
	case "ACTIVE":
		canceller, ok := account.(xenditSessionCanceller)
		if !ok {
			return fmt.Errorf("%w: pembatalan gateway tidak tersedia", utils.ErrConflict)
		}
		_, err = canceller.CancelCardSession(ctx, sessionID, reference)
		return err
	default:
		return fmt.Errorf("%w: sesi gateway sudah diproses; periksa statusnya", utils.ErrConflict)
	}
}

func (s *CheckoutService) cancelTrialCard(ctx context.Context, token string) error {
	r, err := s.trialRecord(ctx, token)
	if err != nil {
		return err
	}
	if r.Status == "cancelled" {
		return nil
	}
	staleCreating := false
	if r.Status == "creating" {
		// A card-save request has a 15-second client timeout. Once this lease is
		// stale and no provider session ID was stored, no checkout SDK was sent
		// to the browser, so the local enrollment can safely be abandoned.
		if err = s.db.GetContext(ctx, &staleCreating, `SELECT provider_session_id IS NULL AND COALESCE(consented_at<=NOW()-INTERVAL '2 minutes',FALSE) FROM trial_card_sessions WHERE id=$1 AND status='creating'`, r.ID); err != nil {
			return err
		}
	}
	if r.Status != "created" && r.Status != "active" && !(r.Status == "creating" && staleCreating) {
		return fmt.Errorf("%w: pendaftaran kartu sedang diproses atau sudah selesai", utils.ErrConflict)
	}
	if r.Provider == "xendit" && r.Status == "active" {
		if r.SessionID == "" {
			return fmt.Errorf("%w: status sesi gateway belum pasti", utils.ErrConflict)
		}
		if err := s.cancelXenditSession(ctx, r.BusinessID, r.SessionID, r.Reference, "SAVE", 0); err != nil {
			return err
		}
	}
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var status string
	var recoverable bool
	if err = tx.QueryRowxContext(ctx, `SELECT status,(provider_session_id IS NULL AND COALESCE(consented_at<=NOW()-INTERVAL '2 minutes',FALSE)) FROM trial_card_sessions WHERE id=$1 FOR UPDATE`, r.ID).Scan(&status, &recoverable); err != nil {
		return err
	}
	if status == "cancelled" {
		return tx.Commit()
	}
	if status != r.Status || (status == "creating" && !recoverable) {
		return fmt.Errorf("%w: status checkout berubah; periksa kembali", utils.ErrConflict)
	}
	var subscriptionStatus string
	if err = tx.GetContext(ctx, &subscriptionStatus, `SELECT status FROM subscriptions WHERE id=$1 FOR UPDATE`, r.SubscriptionID); err != nil {
		return err
	}
	if subscriptionStatus != "pending" {
		return fmt.Errorf("%w: trial sudah dimulai", utils.ErrConflict)
	}
	if _, err = tx.ExecContext(ctx, `UPDATE trial_card_sessions SET status='cancelled',expires_at=NOW() WHERE id=$1`, r.ID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE subscriptions SET status='cancelled' WHERE id=$1 AND status='pending'`, r.SubscriptionID); err != nil {
		return err
	}
	if err = releaseCheckoutCoupons(ctx, tx, r.SubscriptionID); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *CheckoutService) cancelInvoiceCheckout(ctx context.Context, token string) error {
	r, err := s.findByToken(ctx, token)
	if err != nil {
		return err
	}
	if r.Status == "cancelled" {
		return nil
	}
	if r.InvoiceStatus != "pending" || (r.Status != "created" && r.Status != "failed" && !(r.Status == "active" && r.Provider == "xendit")) {
		return fmt.Errorf("%w: pembayaran sedang diproses atau sudah selesai; periksa statusnya", utils.ErrConflict)
	}
	if r.Status == "active" {
		if r.ProviderSessionID == nil || *r.ProviderSessionID == "" || r.PaymentLinkURL != nil {
			return fmt.Errorf("%w: status pembayaran belum pasti", utils.ErrConflict)
		}
		if err := s.cancelXenditSession(ctx, r.BusinessID, *r.ProviderSessionID, r.ReferenceID, "PAY", int64(r.InvoiceAmount)); err != nil {
			return err
		}
	}
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var invoiceStatus string
	if err = tx.GetContext(ctx, &invoiceStatus, `SELECT status FROM invoices WHERE id=$1 FOR UPDATE`, r.InvoiceID); err != nil {
		return err
	}
	var status string
	if err = tx.GetContext(ctx, &status, `SELECT status FROM checkout_sessions WHERE id=$1 FOR UPDATE`, r.ID); err != nil {
		return err
	}
	if status == "cancelled" {
		return tx.Commit()
	}
	if status != r.Status || invoiceStatus != "pending" {
		return fmt.Errorf("%w: status pembayaran berubah; periksa kembali", utils.ErrConflict)
	}
	var subscriptionID int
	err = tx.GetContext(ctx, &subscriptionID, `SELECT i.subscription_id FROM merchant_orders o JOIN invoices i ON i.id=o.invoice_id WHERE o.invoice_id=$1`, r.InvoiceID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if err == nil {
		var subscriptionStatus string
		if err = tx.GetContext(ctx, &subscriptionStatus, `SELECT status FROM subscriptions WHERE id=$1 FOR UPDATE`, subscriptionID); err != nil {
			return err
		}
		if subscriptionStatus != "pending" {
			return fmt.Errorf("%w: langganan tidak lagi menunggu pembayaran", utils.ErrConflict)
		}
		if _, err = tx.ExecContext(ctx, `UPDATE invoices SET status='cancelled' WHERE id=$1 AND status='pending'`, r.InvoiceID); err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, `UPDATE subscriptions SET status='cancelled' WHERE id=$1 AND status='pending'`, subscriptionID); err != nil {
			return err
		}
		if err = releaseCheckoutCoupons(ctx, tx, subscriptionID); err != nil {
			return err
		}
		// An already sending email may be delivered; queued messages must not
		// announce an invoice that has just been cancelled.
		if _, err = tx.ExecContext(ctx, `UPDATE invoice_email_deliveries SET status='failed',last_error='checkout_cancelled' WHERE invoice_id=$1 AND status='queued'`, r.InvoiceID); err != nil {
			return err
		}
	}
	if _, err = tx.ExecContext(ctx, `UPDATE checkout_sessions SET status='cancelled',expires_at=(NOW() AT TIME ZONE 'UTC'),updated_at=(NOW() AT TIME ZONE 'UTC') WHERE id=$1`, r.ID); err != nil {
		return err
	}
	if r.PaymentID != nil {
		if _, err = tx.ExecContext(ctx, `UPDATE payments SET status='expired' WHERE id=$1 AND status='pending'`, *r.PaymentID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func releaseCheckoutCoupons(ctx context.Context, tx interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}, subscriptionID int) error {
	_, err := tx.ExecContext(ctx, `UPDATE coupons c SET used_count=GREATEST(0,c.used_count-q.uses) FROM
 (SELECT coupon_id,count(*)::int uses FROM subscription_coupons WHERE subscription_id=$1 GROUP BY coupon_id) q
 WHERE c.id=q.coupon_id`, subscriptionID)
	return err
}
