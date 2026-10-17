package identity

import (
	"context"
	"database/sql"
	"errors"
	"pay-langgan/internal/utils"
)

type PasswordResetJob struct {
	ID        int64  `db:"id"`
	UserID    int    `db:"user_id"`
	Recipient string `db:"recipient"`
	Nonce     []byte `db:"nonce"`
	TokenHash []byte `db:"token_hash"`
	Attempts  int    `db:"attempts"`
}

func (r *AuthRepository) QueuePasswordReset(ctx context.Context, email string, hash, nonce []byte) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var id int
	err = tx.GetContext(ctx, &id, `SELECT id FROM users WHERE lower(email)=$1 FOR UPDATE`, email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	var recent int
	if err = tx.GetContext(ctx, &recent, `SELECT count(*) FROM password_reset_requests WHERE user_id=$1 AND created_at>NOW()-INTERVAL '1 hour'`, id); err != nil {
		return err
	}
	var cooldown bool
	if err = tx.GetContext(ctx, &cooldown, `SELECT EXISTS(SELECT 1 FROM password_reset_requests WHERE user_id=$1 AND created_at>NOW()-INTERVAL '1 minute')`, id); err != nil {
		return err
	}
	if cooldown || recent >= 5 {
		return nil
	}
	// Issuing another link does not revoke older links until a password is changed.
	_, err = tx.ExecContext(ctx, `INSERT INTO password_reset_requests(user_id,recipient,token_hash,nonce) VALUES($1,$2,$3,$4)`, id, email, hash, nonce)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *AuthRepository) ConsumePasswordReset(ctx context.Context, hash []byte, password string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var userID int
	// Lock the user first so two different links for the same account cannot both succeed.
	err = tx.GetContext(ctx, &userID, `SELECT u.id FROM users u JOIN password_reset_requests p ON p.user_id=u.id WHERE p.token_hash=$1 AND p.recipient=lower(u.email) FOR UPDATE OF u`, hash)
	if errors.Is(err, sql.ErrNoRows) {
		return utils.ErrBadRequest
	}
	if err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, `UPDATE password_reset_requests SET used_at=NOW(),nonce=NULL WHERE token_hash=$1 AND user_id=$2 AND used_at IS NULL AND expires_at>NOW()`, hash, userID)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return utils.ErrBadRequest
	}
	if _, err = tx.ExecContext(ctx, `UPDATE users SET password=$1,password_version=password_version+1,updated_at=NOW() WHERE id=$2`, password, userID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE password_reset_requests SET used_at=COALESCE(used_at,NOW()),nonce=NULL,status=CASE WHEN status IN ('queued','sending') THEN 'cancelled' ELSE status END WHERE user_id=$1`, userID); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *AuthRepository) PasswordVersion(ctx context.Context, userID int) (int, error) {
	var version int
	err := r.db.GetContext(ctx, &version, `SELECT password_version FROM users WHERE id=$1`, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, utils.ErrUnauthorized
	}
	return version, err
}

func (r *AuthRepository) ClaimPasswordReset(ctx context.Context) (*PasswordResetJob, error) {
	_, err := r.db.ExecContext(ctx, `UPDATE password_reset_requests SET status='failed',nonce=NULL WHERE status='sending' AND attempts>=3 AND next_attempt_at<=NOW()`)
	if err != nil {
		return nil, err
	}
	_, err = r.db.ExecContext(ctx, `UPDATE password_reset_requests SET status='cancelled',nonce=NULL WHERE status IN ('queued','sending') AND (expires_at<=NOW() OR used_at IS NOT NULL)`)
	if err != nil {
		return nil, err
	}
	job := &PasswordResetJob{}
	err = r.db.GetContext(ctx, job, `UPDATE password_reset_requests SET status='sending',attempts=attempts+1,next_attempt_at=NOW()+INTERVAL '1 minute' WHERE id=(SELECT p.id FROM password_reset_requests p JOIN users u ON u.id=p.user_id WHERE p.status IN ('queued','sending') AND p.used_at IS NULL AND p.expires_at>NOW() AND p.next_attempt_at<=NOW() AND p.attempts<3 AND lower(u.email)=p.recipient ORDER BY p.id FOR UPDATE OF p SKIP LOCKED LIMIT 1) RETURNING id,user_id,recipient,nonce,token_hash,attempts`)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return job, err
}

func (r *AuthRepository) FinishPasswordReset(ctx context.Context, job *PasswordResetJob, failed bool) error {
	status := "sent"
	if failed {
		status = "queued"
		if job.Attempts >= 3 {
			status = "failed"
		}
	}
	_, err := r.db.ExecContext(ctx, `UPDATE password_reset_requests SET status=$1::varchar,next_attempt_at=NOW()+INTERVAL '1 minute',nonce=CASE WHEN $1::varchar IN ('sent','failed') THEN NULL ELSE nonce END WHERE id=$2 AND status='sending' AND attempts=$3`, status, job.ID, job.Attempts)
	return err
}

// Keep short-lived recovery material out of long-term queue history.
func (r *AuthRepository) CleanPasswordResets(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM password_reset_requests WHERE created_at<NOW()-INTERVAL '1 day'`)
	return err
}
