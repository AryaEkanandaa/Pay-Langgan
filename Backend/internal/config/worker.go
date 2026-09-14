package config

import (
	"fmt"
	"strconv"
	"time"
)

type BillingWorker struct {
	Interval time.Duration
	Timeout  time.Duration
	Limit    int
}

// Call Load first to load Backend/.env. Invalid worker settings fail startup.
func LoadBillingWorker() (BillingWorker, error) {
	var cfg BillingWorker
	var err error
	cfg.Interval, err = time.ParseDuration(getEnv("BILLING_INTERVAL", "1m"))
	if err != nil || cfg.Interval <= 0 {
		return cfg, fmt.Errorf("BILLING_INTERVAL must be a positive duration")
	}
	cfg.Timeout, err = time.ParseDuration(getEnv("BILLING_BATCH_TIMEOUT", "30s"))
	if err != nil || cfg.Timeout <= 0 {
		return cfg, fmt.Errorf("BILLING_BATCH_TIMEOUT must be a positive duration")
	}
	cfg.Limit, err = strconv.Atoi(getEnv("BILLING_BATCH_LIMIT", "100"))
	if err != nil || cfg.Limit < 1 || cfg.Limit > 10000 {
		return cfg, fmt.Errorf("BILLING_BATCH_LIMIT must be 1..10000")
	}
	return cfg, nil
}
