package portal

import (
	"context"
	"crypto/hmac"
	"database/sql"
	"errors"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"pay-langgan/internal/services/mailer"
)

func (s *Service) RunEmailBatch(ctx context.Context) error {
	if s.cfg.EmailDeliveryMode == "disabled" {
		return nil
	}
	base, e := s.baseURL()
	if e != nil {
		return e
	}
	_, e = s.db.ExecContext(ctx, `UPDATE customer_portal_requests r SET status='cancelled',nonce=NULL WHERE status IN ('queued','sending') AND (used_at IS NOT NULL OR expires_at<=NOW() OR NOT EXISTS(SELECT 1 FROM merchant_embed_sites es JOIN businesses b ON b.id=es.business_id JOIN customers c ON c.business_id=b.id WHERE es.id=r.site_id AND es.revoked_at IS NULL AND b.status='active' AND NOT b.deleted AND lower(trim(c.email))=r.recipient));
 UPDATE customer_portal_requests SET status='failed',nonce=NULL WHERE status IN ('queued','sending') AND attempts>=3 AND next_attempt_at<=NOW();
 DELETE FROM customer_portal_sessions WHERE expires_at<=NOW();
 DELETE FROM portal_checkout_access WHERE expires_at<=NOW();
 DELETE FROM customer_portal_requests WHERE expires_at<NOW()-INTERVAL '30 days'`)
	if e != nil {
		return e
	}
	for i := 0; i < 20; i++ {
		var job struct {
			ID        int64  `db:"id"`
			Recipient string `db:"recipient"`
			Nonce     []byte `db:"nonce"`
			Hash      []byte `db:"token_hash"`
			Attempts  int    `db:"attempts"`
			Site      string `db:"public_id"`
			Business  string `db:"name"`
		}
		e = s.db.GetContext(ctx, &job, `WITH claimed AS (UPDATE customer_portal_requests SET status='sending',attempts=attempts+1,next_attempt_at=NOW()+INTERVAL '1 minute' WHERE id=(SELECT id FROM customer_portal_requests WHERE status IN ('queued','sending') AND used_at IS NULL AND expires_at>NOW() AND attempts<3 AND next_attempt_at<=NOW() ORDER BY id FOR UPDATE SKIP LOCKED LIMIT 1) RETURNING *) SELECT r.id,r.recipient,r.nonce,r.token_hash,r.attempts,es.public_id,b.name FROM claimed r JOIN merchant_embed_sites es ON es.id=r.site_id JOIN businesses b ON b.id=es.business_id`)
		if errors.Is(e, sql.ErrNoRows) {
			return nil
		}
		if e != nil {
			return e
		}
		token := magicToken(s.cfg.JWTSecret, job.Nonce)
		status := "sent"
		if len(job.Nonce) != 32 || !hmac.Equal(tokenHash(token), job.Hash) {
			status = "failed"
		} else {
			link := base + "/portal/" + job.Site + "#token=" + token
			body := `<!doctype html><html lang="id"><head><meta charset="utf-8"></head><body style="background:#f0f3fb;padding:24px;font-family:Arial,sans-serif;color:#172033"><main style="max-width:520px;background:white;padding:32px;margin:auto;border-radius:16px;border-top:4px solid #3654ff"><h2>PayLanggan</h2><h1 style="font-size:24px">Portal pelanggan ` + html.EscapeString(job.Business) + `</h1><p>Lihat langganan, invoice, dan lanjutkan pembayaran Anda.</p><p><a href="` + html.EscapeString(link) + `" style="display:inline-block;background:#3654ff;color:white;padding:14px 24px;border-radius:24px;text-decoration:none">Masuk ke portal pelanggan</a></p><p>Tautan berlaku 15 menit sejak permintaan dan hanya dapat digunakan sekali. Jangan bagikan tautan ini.</p><p>Jika Anda tidak meminta tautan ini, abaikan email ini.</p><p style="word-break:break-all;font-size:12px">` + html.EscapeString(link) + `</p></main></body></html>`
			if s.cfg.EmailDeliveryMode == "preview" {
				status = "previewed"
				e = os.MkdirAll(s.cfg.EmailPreviewDir, 0700)
				if e == nil {
					e = os.WriteFile(filepath.Join(s.cfg.EmailPreviewDir, fmt.Sprintf("portal-%d.html", job.ID)), []byte(body), 0600)
				}
			} else {
				e = mailer.Send(ctx, s.cfg, job.Recipient, "", "Tautan portal pelanggan", body, fmt.Sprintf("portal-%d", job.ID))
			}
			if e != nil {
				status = "queued"
				if job.Attempts >= 3 {
					status = "failed"
				}
			}
		}
		_, e = s.db.ExecContext(ctx, `UPDATE customer_portal_requests SET status=$1::varchar,nonce=CASE WHEN $1::varchar IN ('sent','previewed','failed') THEN NULL ELSE nonce END WHERE id=$2 AND status='sending' AND attempts=$3`, status, job.ID, job.Attempts)
		if e != nil {
			return e
		}
	}
	return nil
}
