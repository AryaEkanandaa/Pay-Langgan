package billing

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type RenewalRunner interface {
	Run(context.Context, time.Time, int) (RenewalResult, error)
}

type Scheduler struct {
	runner   RenewalRunner
	logger   *slog.Logger
	interval time.Duration
	timeout  time.Duration
	limit    int
}

func NewScheduler(runner RenewalRunner, logger *slog.Logger, interval, timeout time.Duration, limit int) (*Scheduler, error) {
	if runner == nil || logger == nil || interval <= 0 || timeout <= 0 || limit < 1 || limit > 10000 {
		return nil, fmt.Errorf("scheduler requires runner, logger, positive interval/timeout, and limit 1..10000")
	}
	return &Scheduler{runner: runner, logger: logger, interval: interval, timeout: timeout, limit: limit}, nil
}

// Run starts immediately and waits one interval AFTER each batch completes.
// A single scheduler never overlaps batches; database locks protect other workers.
func (s *Scheduler) Run(ctx context.Context) {
	s.logger.Info("billing_worker_started", "interval", s.interval.String(), "timeout", s.timeout.String(), "limit", s.limit)
	defer s.logger.Info("billing_worker_stopped")
	for ctx.Err() == nil {
		start := time.Now().UTC()
		batchCtx, cancel := context.WithTimeout(ctx, s.timeout)
		result, err := s.runner.Run(batchCtx, start, s.limit)
		cancel()
		attrs := []any{"cutoff", start, "duration_ms", time.Since(start).Milliseconds(), "invoices_created", result.InvoicesCreated, "legacy_skipped", result.LegacySkipped, "failure_count", len(result.Failures)}
		if ctx.Err() != nil {
			s.logger.Info("billing_batch_interrupted", attrs...)
			return
		}
		if err != nil {
			s.logger.Error("billing_batch_failed", append(attrs, "error", err.Error())...)
		} else if len(result.Failures) > 0 {
			s.logger.Warn("billing_batch_partial", attrs...)
		} else {
			s.logger.Info("billing_batch_completed", attrs...)
		}
		for _, failure := range result.Failures {
			s.logger.Error("billing_subscription_failed", "subscription_id", failure.SubscriptionID, "error", failure.Error)
		}
		timer := time.NewTimer(s.interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}
