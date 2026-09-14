package config

import (
	"testing"
	"time"
)

func TestBillingWorkerConfig(t *testing.T) {
	for _, tt := range []struct {
		name, interval, timeout, limit string
		invalid                        bool
	}{
		{"defaults", "", "", "", false},
		{"custom", "5s", "1m", "20", false},
		{"invalid interval", "nope", "", "", true},
		{"zero interval", "0s", "", "", true},
		{"negative timeout", "", "-1s", "", true},
		{"invalid timeout", "", "bad", "", true},
		{"zero limit", "", "", "0", true},
		{"excessive limit", "", "", "10001", true},
		{"invalid limit", "", "", "abc", true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("BILLING_INTERVAL", tt.interval)
			t.Setenv("BILLING_BATCH_TIMEOUT", tt.timeout)
			t.Setenv("BILLING_BATCH_LIMIT", tt.limit)
			cfg, err := LoadBillingWorker()
			if (err != nil) != tt.invalid {
				t.Fatalf("config=%+v, error=%v", cfg, err)
			}
			if tt.name == "defaults" && (cfg.Interval != time.Minute || cfg.Timeout != 30*time.Second || cfg.Limit != 100) {
				t.Fatalf("unexpected defaults: %+v", cfg)
			}
		})
	}
}
