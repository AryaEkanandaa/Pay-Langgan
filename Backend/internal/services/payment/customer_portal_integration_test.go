package payment_test

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/labstack/echo/v4"
	"net/http/httptest"
	"os"
	"path/filepath"
	"pay-langgan/internal/config"
	portalhttp "pay-langgan/internal/handlers/portal"
	"pay-langgan/internal/models"
	billingrepo "pay-langgan/internal/repositories/billing"
	paymentrepo "pay-langgan/internal/repositories/payment"
	"pay-langgan/internal/services/billing"
	paymentservice "pay-langgan/internal/services/payment"
	"pay-langgan/internal/services/portal"
	"pay-langgan/internal/utils"
	provider "pay-langgan/pkg/payment"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestCustomerPortalPostgres(t *testing.T) {
	db := paymentTestDB(t)
	ctx := context.Background()
	exec := func(query string, args ...interface{}) {
		t.Helper()
		if _, e := db.Exec(query, args...); e != nil {
			t.Fatal(e)
		}
	}
	exec(`INSERT INTO businesses(id,name) VALUES('other','Other');
 INSERT INTO merchant_embed_sites(id,business_id,name,public_id,origin) VALUES(1,'test','Portal','pl_site_test',''),(2,'other','Other','pl_site_other','');
 INSERT INTO customers(id,business_id,name,email) VALUES(2,'test','Alya duplicate','ALYA@example.com'),(3,'test','Other customer','other@example.com'),(4,'other','Other tenant','alya@example.com');
 INSERT INTO subscriptions(id,customer_id,plan_id,status,start_date) VALUES(2,2,1,'trial',NOW()),(3,3,1,'active',NOW()),(4,4,1,'active',NOW());
 INSERT INTO invoices(id,subscription_id,invoice_number,amount,status) VALUES(2,2,'INV-DUP',150000,'pending'),(3,3,'INV-PRIVATE',150000,'pending'),(4,4,'INV-OTHER-TENANT',150000,'pending');`)
	cfg := config.Load()
	cfg.EmailDeliveryMode = "preview"
	cfg.EmailPreviewDir = t.TempDir()
	cfg.AppPublicURL = "https://portal.example.invalid"
	svc := portal.NewService(db, cfg)
	if e := svc.RequestLink(ctx, "pl_site_test", "not-an-email"); !errors.Is(e, utils.ErrBadRequest) {
		t.Fatal("invalid email", e)
	}
	if e := svc.RequestLink(ctx, "pl_site_test", "unknown@example.com"); e != nil {
		t.Fatal(e)
	}
	var count int
	db.Get(&count, `SELECT count(*) FROM customer_portal_requests`)
	if count != 0 {
		t.Fatal("unknown customer queued")
	}
	for i := 0; i < 3; i++ {
		if e := svc.RequestLink(ctx, "pl_site_test", "ALYA@example.com"); e != nil {
			t.Fatal(e)
		}
	}
	db.Get(&count, `SELECT count(*) FROM customer_portal_requests`)
	if count != 1 {
		t.Fatal("cooldown or duplicate customer lookup", count)
	}
	if e := svc.RunEmailBatch(ctx); e != nil {
		t.Fatal(e)
	}
	content, e := os.ReadFile(filepath.Join(cfg.EmailPreviewDir, "portal-1.html"))
	if e != nil {
		t.Fatal(e)
	}
	match := regexp.MustCompile(`#token=([A-Za-z0-9_-]{43})`).FindStringSubmatch(string(content))
	if len(match) != 2 {
		t.Fatal("missing fragment link")
	}
	magic := match[1]
	if _, e = svc.Exchange(ctx, "pl_site_other", magic); !errors.Is(e, utils.ErrUnauthorized) {
		t.Fatal("cross-tenant link allowed", e)
	}
	login, e := svc.Exchange(ctx, "pl_site_test", magic)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = svc.Exchange(ctx, "pl_site_test", magic); !errors.Is(e, utils.ErrUnauthorized) {
		t.Fatal("link replay allowed", e)
	}
	v, e := svc.Authenticate(ctx, "pl_site_test", login.Token)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = svc.Authenticate(ctx, "pl_site_other", login.Token); !errors.Is(e, utils.ErrUnauthorized) {
		t.Fatal("cross-site session allowed", e)
	}
	if _, e = svc.Authenticate(ctx, "pl_site_test", magic); !errors.Is(e, utils.ErrUnauthorized) {
		t.Fatal("email token accepted as session", e)
	}
	view, e := svc.Overview(ctx, v, 1, 1)
	if e != nil || len(view.Subscriptions) != 2 || len(view.Invoices) != 2 {
		t.Fatalf("customer scoping %+v %v", view, e)
	}
	if e = svc.OwnInvoice(ctx, v, 3); !errors.Is(e, utils.ErrNotFound) {
		t.Fatal("another customer invoice allowed", e)
	}
	if e = svc.OwnInvoice(ctx, v, 4); !errors.Is(e, utils.ErrNotFound) {
		t.Fatal("another tenant invoice allowed", e)
	}
	xendit, e := provider.NewXenditProvider("", "callback-secret", "https://api.xendit.co", true)
	if e != nil {
		t.Fatal(e)
	}
	checkout, e := paymentservice.NewCheckoutService(db, paymentrepo.NewCheckoutRepository(db), xendit, cfg.AppPublicURL, cfg.AppPublicURL, time.Hour)
	if e != nil {
		t.Fatal(e)
	}
	h := &portalhttp.Handler{Service: svc, Invoice: billing.NewInvoiceService(db, billingrepo.NewInvoiceRepository(db)), Checkout: checkout}
	server := echo.New()
	g := server.Group("/api/v1/portal/:site", h.Headers)
	g.GET("/overview", h.Overview)
	g.POST("/invoices/:id/checkout", h.Pay)
	g.GET("/invoices/:id", h.InvoiceDetail)
	call := func(method, path, token string) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest(method, "/api/v1/portal/pl_site_test"+path, nil)
		r.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		server.ServeHTTP(w, r)
		return w
	}
	if w := call("GET", "/overview", "fake-admin-jwt"); w.Code != 401 {
		t.Fatal("invalid portal credential accepted", w.Code)
	}
	if w := call("GET", "/invoices/3", login.Token); w.Code != 404 {
		t.Fatal("detail ownership bypass", w.Code)
	}
	if w := call("POST", "/invoices/3/checkout", login.Token); w.Code != 404 {
		t.Fatal("checkout ownership bypass", w.Code)
	}
	first := call("POST", "/invoices/1/checkout", login.Token)
	if first.Code != 200 {
		t.Fatal("create checkout", first.Code, first.Body.String())
	}
	var envelope struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if e = json.Unmarshal(first.Body.Bytes(), &envelope); e != nil {
		t.Fatal(e)
	}
	original := envelope.Data.Token
	if _, e = checkout.Start(ctx, original, models.StartCheckoutRequest{Name: "Alya", Email: "alya@example.com", Mobile: "+628123456789"}); e != nil {
		t.Fatal(e)
	}
	second := call("POST", "/invoices/1/checkout", login.Token)
	if second.Code != 200 {
		t.Fatal("resume checkout", second.Code, second.Body.String())
	}
	json.Unmarshal(second.Body.Bytes(), &envelope)
	if envelope.Data.Token == original {
		t.Fatal("resume alias not created")
	}
	if _, e = checkout.Get(ctx, envelope.Data.Token); e != nil {
		t.Fatal("resume token invalid", e)
	}
	if _, e = checkout.Get(ctx, original); e != nil {
		t.Fatal("original checkout invalidated", e)
	}
	db.Get(&count, `SELECT count(*) FROM payments WHERE invoice_id=1`)
	if count != 1 {
		t.Fatal("resume created second transaction", count)
	}
	exec(`UPDATE invoices SET status='paid' WHERE id=1`)
	if w := call("POST", "/invoices/1/checkout", login.Token); w.Code != 409 {
		t.Fatal("paid invoice allowed checkout", w.Code)
	}
	if e = svc.Logout(ctx, login.Token); e != nil {
		t.Fatal(e)
	}
	if _, e = svc.Authenticate(ctx, "pl_site_test", login.Token); !errors.Is(e, utils.ErrUnauthorized) {
		t.Fatal("logout did not revoke session", e)
	}
	// Fresh link: expired links and revoked integration IDs must fail.
	exec(`UPDATE customer_portal_requests SET created_at=NOW()-INTERVAL '2 minutes'`)
	if e = svc.RequestLink(ctx, "pl_site_test", "alya@example.com"); e != nil {
		t.Fatal(e)
	}
	if e = svc.RunEmailBatch(ctx); e != nil {
		t.Fatal(e)
	}
	content, e = os.ReadFile(filepath.Join(cfg.EmailPreviewDir, "portal-2.html"))
	if e != nil {
		t.Fatal(e)
	}
	magic = regexp.MustCompile(`#token=([A-Za-z0-9_-]{43})`).FindStringSubmatch(string(content))[1]
	exec(`UPDATE customer_portal_requests SET expires_at=NOW()-INTERVAL '1 minute' WHERE id=2`)
	if _, e = svc.Exchange(ctx, "pl_site_test", magic); !errors.Is(e, utils.ErrUnauthorized) {
		t.Fatal("expired link allowed", e)
	}
	exec(`UPDATE customer_portal_requests SET expires_at=NOW()+INTERVAL '1 minute' WHERE id=2`)
	login, e = svc.Exchange(ctx, "pl_site_test", magic)
	if e != nil {
		t.Fatal(e)
	}
	exec(`UPDATE customer_portal_sessions SET expires_at=NOW()-INTERVAL '1 second'`)
	if _, e = svc.Authenticate(ctx, "pl_site_test", login.Token); !errors.Is(e, utils.ErrUnauthorized) {
		t.Fatal("expired session allowed", e)
	}
	exec(`UPDATE customer_portal_sessions SET expires_at=NOW()+INTERVAL '1 hour'; UPDATE merchant_embed_sites SET revoked_at=NOW() WHERE id=1`)
	if _, e = svc.Authenticate(ctx, "pl_site_test", login.Token); !errors.Is(e, utils.ErrUnauthorized) {
		t.Fatal("revoked integration session allowed", e)
	}
	if strings.Contains(string(content), "localhost") {
		t.Fatal("email contains local URL")
	}
}
