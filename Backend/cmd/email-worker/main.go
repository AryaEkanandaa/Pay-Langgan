package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"pay-langgan/internal/config"
	"pay-langgan/internal/database"
	repo "pay-langgan/internal/repositories/billing"
	identityrepo "pay-langgan/internal/repositories/identity"
	"pay-langgan/internal/services/billing"
	"pay-langgan/internal/services/identity"
	"pay-langgan/internal/services/portal"
	"time"
)

func main() {
	cfg := config.Load()
	if err := billing.ValidateInvoiceEmailConfig(cfg); err != nil {
		log.Fatal(err)
	}
	db, err := database.New(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	svc := billing.NewInvoiceService(db, repo.NewInvoiceRepository(db))
	auth := identity.NewAuthService(identityrepo.NewAuthRepository(db), cfg)
	customerPortal := portal.NewService(db, cfg)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	log.Printf("Invoice email worker: %s", cfg.EmailDeliveryMode)
	for {
		if cfg.EmailDeliveryMode != "disabled" {
			if err := svc.QueueDueReminders(ctx, time.Now()); err != nil {
				log.Print("Email reminder scheduling failed")
			}
		}
		if err := auth.RunPasswordResetBatch(ctx); err != nil {
			log.Print("Password reset email queue processing failed")
		}
		if err := customerPortal.RunEmailBatch(ctx); err != nil {
			log.Print("Customer portal email queue processing failed")
		}
		if err := svc.RunEmailBatch(ctx, cfg); err != nil {
			log.Print("Email queue processing failed")
		}
		if err := svc.RunCheckoutEmailBatch(ctx, cfg); err != nil {
			log.Print("Checkout confirmation email queue processing failed")
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Second):
		}
	}
}
