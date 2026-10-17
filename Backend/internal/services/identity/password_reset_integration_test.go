package identity_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
	"pay-langgan/internal/config"
	"pay-langgan/internal/database"
	handler "pay-langgan/internal/handlers/identity"
	"pay-langgan/internal/middlewares"
	"pay-langgan/internal/models"
	repo "pay-langgan/internal/repositories/identity"
	service "pay-langgan/internal/services/identity"
	"pay-langgan/internal/utils"
)

func resetTestService(t *testing.T) (*service.AuthService, *database.DB, *config.Config) {
	t.Helper()
	if os.Getenv("AUTH_INTEGRATION") != "1" {
		t.Skip("set AUTH_INTEGRATION=1 for isolated PostgreSQL tests")
	}
	_ = godotenv.Load("../../../.env")
	cfg := config.Load()
	admin, err := database.New(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)
	if err != nil {
		t.Fatal(err)
	}
	schema := fmt.Sprintf("reset_test_%d", time.Now().UnixNano())
	if _, err = admin.Exec("CREATE SCHEMA " + pq.QuoteIdentifier(schema)); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { admin.Exec("DROP SCHEMA " + pq.QuoteIdentifier(schema) + " CASCADE"); admin.Close() })
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s search_path=%s", cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode, schema)
	pool, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { pool.Close() })
	db := &database.DB{DB: pool}
	files, err := filepath.Glob("../../../../Migrations/*.sql")
	if err != nil || len(files) == 0 {
		t.Fatal("missing migrations")
	}
	for _, file := range files {
		data, e := os.ReadFile(file)
		if e != nil {
			t.Fatal(e)
		}
		if _, e = db.Exec(strings.Split(string(data), "-- +goose Down")[0]); e != nil {
			t.Fatalf("migration %s: %v", file, e)
		}
	}
	hash, _ := utils.HashPassword("old-password")
	if _, err = db.Exec(`INSERT INTO businesses(id,name) VALUES('reset-test','Reset Test')`); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`INSERT INTO users(id,business_id,name,email,password,role) VALUES(1,'reset-test','Admin','admin@example.invalid',$1,'admin'),(2,'reset-test','Other','other@example.invalid',$1,'admin')`, hash); err != nil {
		t.Fatal(err)
	}
	cfg.EmailDeliveryMode = "preview"
	cfg.EmailPreviewDir = t.TempDir()
	cfg.AppPublicURL = "https://frontend.paylanggan.example"
	cfg.JWTSecret = "test-only-password-reset-secret"
	cfg.JWTExpiresIn = time.Hour
	return service.NewAuthService(repo.NewAuthRepository(db), cfg), db, cfg
}

func resetTokenFromPreview(t *testing.T, cfg *config.Config, id int64) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(cfg.EmailPreviewDir, fmt.Sprintf("password-reset-%d.html", id)))
	if err != nil {
		t.Fatal(err)
	}
	match := regexp.MustCompile(`/reset-password#token=([A-Za-z0-9_-]{43})`).FindStringSubmatch(string(data))
	if len(match) != 2 {
		t.Fatal("preview lacks safe reset link")
	}
	return match[1]
}

func TestPasswordResetPostgres(t *testing.T) {
	svc, db, cfg := resetTestService(t)
	ctx := context.Background()
	exec := func(q string, args ...interface{}) {
		t.Helper()
		if _, err := db.Exec(q, args...); err != nil {
			t.Fatal(err)
		}
	}
	before, err := svc.Login(models.LoginRequest{Email: "admin@example.invalid", Password: "old-password"})
	if err != nil {
		t.Fatal(err)
	}
	if err = svc.RequestPasswordReset(ctx, "missing@example.invalid"); err != nil {
		t.Fatal("unknown email response differs", err)
	}
	var count int
	db.Get(&count, `SELECT count(*) FROM password_reset_requests`)
	if count != 0 {
		t.Fatal("unknown email queued")
	}
	if err = svc.RequestPasswordReset(ctx, " ADMIN@example.invalid "); err != nil {
		t.Fatal(err)
	}
	if err = svc.RequestPasswordReset(ctx, "admin@example.invalid"); err != nil {
		t.Fatal(err)
	}
	db.Get(&count, `SELECT count(*) FROM password_reset_requests`)
	if count != 1 {
		t.Fatal("cooldown failed")
	}
	var id int64
	db.Get(&id, `SELECT id FROM password_reset_requests`)
	var expires time.Time
	db.Get(&expires, `SELECT expires_at FROM password_reset_requests`)
	if time.Until(expires) > 15*time.Minute || time.Until(expires) < 14*time.Minute {
		t.Fatal("wrong expiry")
	}
	if err = svc.RunPasswordResetBatch(ctx); err != nil {
		t.Fatal(err)
	}
	token := resetTokenFromPreview(t, cfg, id)
	var nonce []byte
	db.Get(&nonce, `SELECT nonce FROM password_reset_requests`)
	if len(nonce) != 0 {
		t.Fatal("recovery material retained after delivery")
	}
	if err = svc.ResetPassword(ctx, token, "new-password", "wrong"); !errors.Is(err, utils.ErrBadRequest) {
		t.Fatal("confirmation mismatch accepted", err)
	}
	exec(`UPDATE password_reset_requests SET expires_at=NOW()-INTERVAL '1 second'`)
	if err = svc.ResetPassword(ctx, token, "new-password", "new-password"); !errors.Is(err, utils.ErrBadRequest) {
		t.Fatal("expired link accepted", err)
	}
	exec(`UPDATE password_reset_requests SET created_at=NOW()-INTERVAL '2 minutes'`)
	if err = svc.RequestPasswordReset(ctx, "admin@example.invalid"); err != nil {
		t.Fatal(err)
	}
	db.Get(&id, `SELECT max(id) FROM password_reset_requests`)
	if err = svc.RunPasswordResetBatch(ctx); err != nil {
		t.Fatal(err)
	}
	token = resetTokenFromPreview(t, cfg, id)
	if err = svc.ResetPassword(ctx, token, "new-password", "new-password"); err != nil {
		t.Fatal(err)
	}
	if err = svc.ResetPassword(ctx, token, "another-password", "another-password"); !errors.Is(err, utils.ErrBadRequest) {
		t.Fatal("replay accepted", err)
	}
	if _, err = svc.Login(models.LoginRequest{Email: "admin@example.invalid", Password: "old-password"}); !errors.Is(err, utils.ErrUnauthorized) {
		t.Fatal("old password accepted", err)
	}
	after, err := svc.Login(models.LoginRequest{Email: "admin@example.invalid", Password: "new-password"})
	if err != nil {
		t.Fatal(err)
	}
	oldClaims, _ := utils.ParseToken(cfg.JWTSecret, before.Token)
	newClaims, _ := utils.ParseToken(cfg.JWTSecret, after.Token)
	if err = svc.VerifySession(ctx, 1, oldClaims.PasswordVersion); !errors.Is(err, utils.ErrUnauthorized) {
		t.Fatal("old JWT not revoked", err)
	}
	if err = svc.VerifySession(ctx, 1, newClaims.PasswordVersion); err != nil {
		t.Fatal("new JWT invalid", err)
	}
	if _, err = svc.Login(models.LoginRequest{Email: "other@example.invalid", Password: "old-password"}); err != nil {
		t.Fatal("another account changed", err)
	}

	// Two different valid links race: only one may change this account's password.
	var tokens []string
	for i := 0; i < 2; i++ {
		exec(`UPDATE password_reset_requests SET created_at=NOW()-INTERVAL '2 minutes'`)
		if err = svc.RequestPasswordReset(ctx, "admin@example.invalid"); err != nil {
			t.Fatal(err)
		}
		db.Get(&id, `SELECT max(id) FROM password_reset_requests`)
		if err = svc.RunPasswordResetBatch(ctx); err != nil {
			t.Fatal(err)
		}
		tokens = append(tokens, resetTokenFromPreview(t, cfg, id))
	}
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for _, value := range tokens {
		wg.Add(1)
		go func(token string) {
			defer wg.Done()
			results <- svc.ResetPassword(ctx, token, "race-password", "race-password")
		}(value)
	}
	wg.Wait()
	close(results)
	success := 0
	for err := range results {
		if err == nil {
			success++
		} else if !errors.Is(err, utils.ErrBadRequest) {
			t.Fatal(err)
		}
	}
	if success != 1 {
		t.Fatal("multiple links succeeded concurrently", success)
	}
}

func TestPasswordResetWorkerRetryPostgres(t *testing.T) {
	svc, db, cfg := resetTestService(t)
	ctx := context.Background()
	if err := svc.RequestPasswordReset(ctx, "admin@example.invalid"); err != nil {
		t.Fatal(err)
	}
	cfg.EmailPreviewDir = filepath.Join(cfg.EmailPreviewDir, "blocked")
	if err := os.WriteFile(cfg.EmailPreviewDir, []byte("block"), 0600); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		if err := svc.RunPasswordResetBatch(ctx); err != nil {
			t.Fatal(err)
		}
		db.Exec(`UPDATE password_reset_requests SET next_attempt_at=NOW()`)
	}
	var status string
	db.Get(&status, `SELECT status FROM password_reset_requests`)
	if status != "failed" {
		t.Fatal("final retry not failed", status)
	}
	var nonce []byte
	db.Get(&nonce, `SELECT nonce FROM password_reset_requests`)
	if len(nonce) != 0 {
		t.Fatal("failed recovery material retained")
	}
}

func TestPasswordResetHTTPPostgres(t *testing.T) {
	svc, db, cfg := resetTestService(t)
	h := handler.NewAuthHandler(svc)
	e := echo.New()
	e.POST("/forgot", h.ForgotPassword, h.PasswordResetLimit())
	e.POST("/reset", h.ResetPassword, h.PasswordResetLimit())
	e.GET("/me", h.Me, middlewares.JWTAuth(cfg.JWTSecret, func(echo.Context) bool { return false }), h.RequireCurrentSession)
	server := httptest.NewServer(e)
	defer server.Close()
	post := func(path string, payload interface{}) (int, string) {
		t.Helper()
		raw, _ := json.Marshal(payload)
		res, err := http.Post(server.URL+path, "application/json", bytes.NewReader(raw))
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		data, _ := io.ReadAll(res.Body)
		return res.StatusCode, string(data)
	}
	knownStatus, known := post("/forgot", map[string]string{"email": "admin@example.invalid"})
	unknownStatus, unknown := post("/forgot", map[string]string{"email": "unknown@example.invalid"})
	if knownStatus != 200 || unknownStatus != 200 || known != unknown {
		t.Fatal("account enumeration in response", knownStatus, unknownStatus)
	}
	for i := 0; i < 3; i++ {
		post("/forgot", map[string]string{"email": "unknown@example.invalid"})
	}
	limited, _ := post("/forgot", map[string]string{"email": "admin@example.invalid"})
	if limited != 429 {
		t.Fatal("rate limiting missing", limited)
	}
	if err := svc.RunPasswordResetBatch(context.Background()); err != nil {
		t.Fatal(err)
	}
	var id int64
	db.Get(&id, `SELECT max(id) FROM password_reset_requests`)
	token := resetTokenFromPreview(t, cfg, id)
	if strings.Contains(known, token) {
		t.Fatal("token leaked in response")
	}
	login, err := svc.Login(models.LoginRequest{Email: "admin@example.invalid", Password: "old-password"})
	if err != nil {
		t.Fatal(err)
	}
	checkSession := func(token string) int {
		t.Helper()
		req, _ := http.NewRequest("GET", server.URL+"/me", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		return res.StatusCode
	}
	if checkSession(login.Token) != 200 {
		t.Fatal("initial session invalid")
	}
	status, _ := post("/reset", map[string]string{"token": token, "password": "new-password", "password_confirmation": "new-password"})
	if status != 200 {
		t.Fatal("HTTP reset failed", status)
	}
	if checkSession(login.Token) != 401 {
		t.Fatal("old HTTP session remained valid")
	}
	fresh, err := svc.Login(models.LoginRequest{Email: "admin@example.invalid", Password: "new-password"})
	if err != nil {
		t.Fatal(err)
	}
	if checkSession(fresh.Token) != 200 {
		t.Fatal("new HTTP session invalid")
	}
	replay, _ := post("/reset", map[string]string{"token": token, "password": "new-password", "password_confirmation": "new-password"})
	if replay != 400 {
		t.Fatal("HTTP token replay allowed", replay)
	}
}
