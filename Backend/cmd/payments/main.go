// Payments is a separately deployed reconciliation and webhook retry worker.
// Manual recovery options are operator-only; none are public HTTP endpoints.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pay-langgan/internal/config"
	"pay-langgan/internal/database"
	repo "pay-langgan/internal/repositories/payment"
	service "pay-langgan/internal/services/payment"
	provider "pay-langgan/pkg/payment"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "payment maintenance failed; inspect checkout_sessions.last_error and webhook_events.last_error")
		os.Exit(1)
	}
}

func run() error {
	once := flag.Bool("once", false, "run one maintenance batch then exit")
	interval := flag.Duration("interval", time.Minute, "delay between batches")
	timeout := flag.Duration("timeout", 2*time.Minute, "maximum duration of a batch or manual operation")
	limit := flag.Int("limit", 50, "maximum webhook events and sessions per batch (1..1000)")
	retryEvent := flag.Int64("retry-event", 0, "retry an unprocessed/dead-letter webhook by ID")
	recoverCheckout := flag.Int64("recover-checkout", 0, "reconcile one checkout by ID")
	sessionID := flag.String("session-id", "", "verified provider session ID for a lost creation response")
	recoverAutopay := flag.Int64("recover-autopay", 0, "reconcile an automatic charge by local charge ID")
	requestID := flag.String("request-id", "", "Xendit payment request ID for automatic charge recovery")
	flag.Parse()
	if *recoverAutopay < 0 || (*recoverAutopay > 0 && (*requestID == "" || *retryEvent != 0 || *recoverCheckout != 0)) || (*requestID != "" && *recoverAutopay == 0) {
		return fmt.Errorf("invalid automatic charge recovery arguments")
	}
	if *timeout <= 0 || *interval <= 0 || *limit < 1 || *limit > 1000 || *retryEvent < 0 || *recoverCheckout < 0 ||
		(*retryEvent != 0 && *recoverCheckout != 0) || (*sessionID != "" && *recoverCheckout == 0) {
		return fmt.Errorf("invalid payment worker arguments")
	}
	cfg := config.Load()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	db, err := database.New(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)
	if err != nil {
		return err
	}
	defer db.Close()
	xendit, err := provider.NewXenditProvider(cfg.XenditSecretKey, cfg.XenditCallbackToken, cfg.XenditBaseURL, cfg.XenditMock)
	if err != nil {
		return err
	}
	checkout, err := service.NewCheckoutService(db, repo.NewCheckoutRepository(db), xendit, cfg.AppPublicURL, cfg.XenditReturnBaseURL, cfg.CheckoutTTL)
	if err != nil {
		return err
	}
	checkout.SetBusinessGatewayResolver(service.NewBusinessGatewayService(db, cfg.PaymentCredentialEncryptionKey, cfg.APIPublicURL))
	checkout.EnableAutopay(cfg.AutopayEnabled)
	if *once || *retryEvent != 0 || *recoverCheckout != 0 || *recoverAutopay != 0 {
		ctx, cancel := context.WithTimeout(ctx, *timeout)
		defer cancel()
		if *recoverAutopay != 0 {
			return checkout.RecoverAutopay(ctx, *recoverAutopay, *requestID)
		}
		if *retryEvent != 0 {
			return checkout.RetryWebhook(ctx, *retryEvent)
		}
		if *recoverCheckout != 0 {
			return checkout.RecoverCheckout(ctx, *recoverCheckout, *sessionID)
		}
		result, err := checkout.RunMaintenance(ctx, *limit)
		if encodeErr := json.NewEncoder(os.Stdout).Encode(result); encodeErr != nil {
			return encodeErr
		}
		if err != nil {
			return err
		}
		if result.Failures > 0 {
			return fmt.Errorf("payment maintenance has failures")
		}
		return nil
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	worker, err := service.NewWorker(checkout, logger, *interval, *timeout, *limit)
	if err != nil {
		return err
	}
	worker.Run(ctx)
	return nil
}
