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
	service "pay-langgan/internal/services/platformbilling"
	provider "pay-langgan/pkg/payment"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "platform billing worker failed:", err)
		os.Exit(1)
	}
}

func run() error {
	once := flag.Bool("once", false, "run one batch and exit")
	limit := flag.Int("limit", 50, "maximum items per phase (1..1000)")
	interval := flag.Duration("interval", time.Minute, "delay between batches")
	timeout := flag.Duration("timeout", 2*time.Minute, "maximum batch duration")
	flag.Parse()
	if *limit < 1 || *limit > 1000 || *interval <= 0 || *timeout <= 0 {
		return fmt.Errorf("invalid worker arguments")
	}
	cfg := config.Load()
	db, err := database.New(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)
	if err != nil {
		return err
	}
	defer db.Close()
	xendit, err := provider.NewXenditProvider(cfg.XenditSecretKey, cfg.XenditCallbackToken, cfg.XenditBaseURL, cfg.XenditMock)
	if err != nil {
		return err
	}
	billing, err := service.NewService(db, xendit, cfg.AppPublicURL, cfg.XenditReturnBaseURL)
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	for {
		batchCtx, cancel := context.WithTimeout(ctx, *timeout)
		result, runErr := billing.RunOnce(batchCtx, *limit)
		cancel()
		if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
			return err
		}
		if runErr != nil {
			logger.Error("platform_billing_batch_failed", "error", runErr)
		}
		if *once {
			if runErr != nil {
				return runErr
			}
			if result.Failures > 0 {
				return fmt.Errorf("batch has %d failures", result.Failures)
			}
			return nil
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(*interval):
		}
	}
}
