package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	"pay-langgan/internal/config"
	"pay-langgan/internal/database"
	"pay-langgan/internal/services/billing"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	atFlag := flag.String("at", "", "billing cutoff in RFC3339 (default: now); writes invoices, not a dry run")
	limit := flag.Int("limit", 100, "maximum periods processed (1..10000)")
	flag.Parse()
	at := time.Now().UTC()
	if *atFlag != "" {
		parsed, err := time.Parse(time.RFC3339Nano, *atFlag)
		if err != nil {
			return fmt.Errorf("invalid -at: %w", err)
		}
		at = parsed.UTC()
	}
	if *limit < 1 || *limit > 10000 {
		return fmt.Errorf("limit must be 1..10000")
	}
	cfg := config.Load()
	db, err := database.New(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)
	if err != nil {
		return err
	}
	defer db.Close()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	result, err := billing.NewRenewalService(db).Run(ctx, at, *limit)
	if encodeErr := json.NewEncoder(os.Stdout).Encode(result); encodeErr != nil {
		return encodeErr
	}
	if err != nil {
		return err
	}
	if len(result.Failures) > 0 {
		return fmt.Errorf("%d subscriptions failed renewal", len(result.Failures))
	}
	return nil
}
