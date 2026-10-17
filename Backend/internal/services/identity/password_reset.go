package identity

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"html"
	"net/mail"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"pay-langgan/internal/services/mailer"
	"pay-langgan/internal/utils"
)

// The queue stores a random nonce, not the bearer token. Only the server with
// JWT_SECRET can reconstruct the HMAC token for delivery. Verification uses SHA256.
func resetToken(secret string, nonce []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte("paylanggan-password-reset-v1:"))
	mac.Write(nonce)
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func (s *AuthService) RequestPasswordReset(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email || len(email) > 100 {
		return utils.ErrBadRequest
	}
	if s.cfg.EmailDeliveryMode != "smtp" && s.cfg.EmailDeliveryMode != "preview" {
		return fmt.Errorf("password reset email is unavailable")
	}
	if _, err = s.resetBaseURL(); err != nil {
		return err
	}
	nonce := make([]byte, 32)
	if _, err = rand.Read(nonce); err != nil {
		return err
	}
	token := resetToken(s.cfg.JWTSecret, nonce)
	hash := sha256.Sum256([]byte(token))
	return s.authRepo.QueuePasswordReset(ctx, email, hash[:], nonce)
}

func (s *AuthService) ResetPassword(ctx context.Context, token, password, confirmation string) error {
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil || len(raw) != 32 || len(token) != 43 {
		return utils.ErrBadRequest
	}
	if password != confirmation || utf8.RuneCountInString(password) < 8 || len(password) > 72 {
		return utils.ErrBadRequest
	}
	hash, err := utils.HashPassword(password)
	if err != nil {
		return err
	}
	tokenHash := sha256.Sum256([]byte(token))
	return s.authRepo.ConsumePasswordReset(ctx, tokenHash[:], hash)
}

func (s *AuthService) VerifySession(ctx context.Context, userID, version int) error {
	current, err := s.authRepo.PasswordVersion(ctx, userID)
	if err != nil {
		return err
	}
	if current != version {
		return utils.ErrUnauthorized
	}
	return nil
}

func (s *AuthService) resetBaseURL() (string, error) {
	u, err := url.Parse(s.cfg.AppPublicURL)
	if err != nil || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return "", fmt.Errorf("invalid APP_PUBLIC_URL")
	}
	local := u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1" || u.Hostname() == "::1"
	if u.Scheme != "https" && !(s.cfg.EmailDeliveryMode == "preview" && local && u.Scheme == "http") {
		return "", fmt.Errorf("password reset requires public HTTPS APP_PUBLIC_URL")
	}
	return strings.TrimRight(u.String(), "/"), nil
}

func (s *AuthService) RunPasswordResetBatch(ctx context.Context) error {
	if s.cfg.EmailDeliveryMode == "disabled" {
		return nil
	}
	base, err := s.resetBaseURL()
	if err != nil {
		return err
	}
	if err = s.authRepo.CleanPasswordResets(ctx); err != nil {
		return err
	}
	for i := 0; i < 20; i++ {
		job, err := s.authRepo.ClaimPasswordReset(ctx)
		if err != nil {
			return err
		}
		if job == nil {
			return nil
		}
		token := resetToken(s.cfg.JWTSecret, job.Nonce)
		hash := sha256.Sum256([]byte(token))
		if len(job.Nonce) != 32 || !hmac.Equal(hash[:], job.TokenHash) {
			if err = s.authRepo.FinishPasswordReset(ctx, job, true); err != nil {
				return err
			}
			continue
		}
		link := base + "/reset-password#token=" + url.QueryEscape(token)
		body := `<!doctype html><html lang="id"><head><meta charset="utf-8"></head><body style="margin:0;padding:32px;background:#f6f7fb;font-family:Arial,sans-serif;color:#172033"><div style="max-width:520px;margin:auto;padding:32px;background:#fff;border-radius:16px;border-top:4px solid #3654ff"><h2>PayLanggan</h2><h1 style="font-size:24px">Atur ulang password</h1><p>Kami menerima permintaan untuk mengatur ulang password akun PayLanggan Anda.</p><p><a href="` + html.EscapeString(link) + `" style="display:inline-block;padding:14px 24px;border-radius:24px;background:#3654ff;color:#fff;text-decoration:none">Buat password baru</a></p><p>Tautan berlaku selama 15 menit sejak permintaan dibuat dan hanya dapat digunakan sekali.</p><p>Jika Anda tidak meminta perubahan ini, abaikan email ini. Password Anda belum berubah.</p><p style="font-size:12px;color:#475569">Jika tombol tidak bekerja, salin alamat berikut ke browser:</p><p style="font-size:12px;word-break:break-all">` + html.EscapeString(link) + `</p></div></body></html>`
		if s.cfg.EmailDeliveryMode == "preview" {
			err = os.MkdirAll(s.cfg.EmailPreviewDir, 0700)
			if err == nil {
				err = os.WriteFile(filepath.Join(s.cfg.EmailPreviewDir, fmt.Sprintf("password-reset-%d.html", job.ID)), []byte(body), 0600)
			}
		} else if s.cfg.EmailDeliveryMode == "smtp" {
			err = mailer.Send(ctx, s.cfg, job.Recipient, "", "Atur ulang password PayLanggan", body, fmt.Sprintf("password-reset-%d", job.ID))
		} else {
			err = fmt.Errorf("invalid email mode")
		}
		if finishErr := s.authRepo.FinishPasswordReset(ctx, job, err != nil); finishErr != nil {
			return finishErr
		}
	}
	return nil
}
