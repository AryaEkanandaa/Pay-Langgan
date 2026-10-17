package payment_test

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	paymentrepo "pay-langgan/internal/repositories/payment"
	paymentservice "pay-langgan/internal/services/payment"
	"pay-langgan/internal/utils"
	provider "pay-langgan/pkg/payment"
)

func TestWebhookUnreadPostgres(t *testing.T) {
	db := paymentTestDB(t)
	xendit, err := provider.NewXenditProvider("", "unused", "https://api.xendit.co", true)
	if err != nil {
		t.Fatal(err)
	}
	service, err := paymentservice.NewCheckoutService(db, paymentrepo.NewCheckoutRepository(db), xendit, "http://localhost:5173", "https://paylanggan.example", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	exec := func(query string, args ...interface{}) {
		t.Helper()
		if _, err := db.Exec(query, args...); err != nil {
			t.Fatal(err)
		}
	}
	exec(`INSERT INTO businesses(id,name) VALUES('other','Other');
 INSERT INTO users(id,business_id,name,email,password) VALUES(1,'test','Admin A','a@example.invalid','unused'),(2,'test','Admin B','b@example.invalid','unused'),(3,'other','Admin C','c@example.invalid','unused');
 UPDATE users SET webhook_logs_last_seen_at=NOW()-INTERVAL '1 hour';
 INSERT INTO business_webhook_logs(business_id,provider,status,http_status,received_at) VALUES('test','midtrans','accepted',200,NOW()-INTERVAL '2 hours'),('test','xendit','failed',401,NOW()-INTERVAL '1 minute'),('other','midtrans','accepted',200,NOW()-INTERVAL '1 minute')`)
	ctx := context.Background()
	initial, err := service.NewWebhookCount(ctx, "test", 1)
	if err != nil || initial.Count != 1 {
		t.Fatalf("tenant count: %+v %v", initial, err)
	}
	if _, err = service.NewWebhookCount(ctx, "other", 1); !errors.Is(err, utils.ErrUnauthorized) {
		t.Fatal("cross-tenant user accepted", err)
	}
	exec(`INSERT INTO business_webhook_logs(business_id,provider,status,http_status,received_at) VALUES('test','midtrans','duplicate',200,$1)`, initial.CheckedAt.Add(time.Microsecond))
	seen, err := service.MarkWebhookLogsSeen(ctx, "test", 1, initial.CheckedAt)
	if err != nil || seen.Count != 1 {
		t.Fatalf("arrival during list load was lost: %+v %v", seen, err)
	}
	otherAdmin, err := service.NewWebhookCount(ctx, "test", 2)
	if err != nil || otherAdmin.Count != 2 {
		t.Fatalf("other admin changed: %+v %v", otherAdmin, err)
	}
	seen, err = service.MarkWebhookLogsSeen(ctx, "test", 1, seen.CheckedAt)
	if err != nil || seen.Count != 0 {
		t.Fatalf("read count: %+v %v", seen, err)
	}
	last := seen.LastSeenAt
	seen, err = service.MarkWebhookLogsSeen(ctx, "test", 1, initial.LastSeenAt)
	if err != nil || !seen.LastSeenAt.Equal(last) {
		t.Fatal("older device moved cursor backwards", err)
	}
	seen, err = service.MarkWebhookLogsSeen(ctx, "test", 1, time.Now().Add(24*time.Hour))
	if err != nil || seen.LastSeenAt.After(seen.CheckedAt) {
		t.Fatal("future cursor accepted", err)
	}
	if _, err = service.MarkWebhookLogsSeen(ctx, "other", 1, initial.CheckedAt); !errors.Is(err, utils.ErrUnauthorized) {
		t.Fatal("cross-tenant update accepted", err)
	}
	if _, err = service.MarkWebhookLogsSeen(ctx, "test", 1, time.Time{}); !errors.Is(err, utils.ErrBadRequest) {
		t.Fatal("empty cursor accepted", err)
	}

	// Reapply only this migration in the disposable schema with existing users/logs.
	exec(`ALTER TABLE users DROP COLUMN webhook_logs_last_seen_at`)
	migration, err := os.ReadFile("../../../../Migrations/2026-10-07-39_webhook_logs_last_seen.sql")
	if err != nil {
		t.Fatal(err)
	}
	exec(strings.Split(string(migration), "-- +goose Down")[0])
	migrated, err := service.NewWebhookCount(ctx, "test", 2)
	if err != nil || migrated.Count != 0 {
		t.Fatalf("old history marked new after migration: %+v %v", migrated, err)
	}
}
