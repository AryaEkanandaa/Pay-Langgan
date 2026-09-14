package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"pay-langgan/internal/config"
	"pay-langgan/internal/database"
	"pay-langgan/internal/services/billing"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("billing_worker_startup_failed", "error", err.Error())
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg := config.Load()
	worker, err := config.LoadBillingWorker()
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	db, err := database.New(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)
	if err != nil {
		return err
	}
	defer db.Close()
	scheduler, err := billing.NewScheduler(billing.NewRenewalService(db), logger, worker.Interval, worker.Timeout, worker.Limit)
	if err != nil {
		return err
	}
	scheduler.Run(ctx)
	return nil
}
