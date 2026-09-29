package payment

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type MaintenanceRunner interface {
	RunMaintenance(context.Context, int) (MaintenanceResult, error)
}

type Worker struct {
	runner   MaintenanceRunner
	logger   *slog.Logger
	interval time.Duration
	timeout  time.Duration
	limit    int
}

func NewWorker(runner MaintenanceRunner, logger *slog.Logger, interval, timeout time.Duration, limit int) (*Worker, error) {
	if runner == nil || logger == nil || interval <= 0 || timeout <= 0 || limit < 1 || limit > 1000 {
		return nil, fmt.Errorf("payment worker requires runner, logger, positive durations and limit 1..1000")
	}
	return &Worker{runner: runner, logger: logger, interval: interval, timeout: timeout, limit: limit}, nil
}

func (w *Worker) Run(ctx context.Context) {
	for ctx.Err() == nil {
		batchCtx, cancel := context.WithTimeout(ctx, w.timeout)
		result, err := w.runner.RunMaintenance(batchCtx, w.limit)
		cancel()
		// Never include provider response bodies, tokens or SDK keys in logs.
		w.logger.Info("payment_maintenance", "webhooks_processed", result.WebhooksProcessed,
			"autopay_started", result.AutopayStarted, "autopay_checked", result.AutopayChecked,
			"sessions_checked", result.SessionsChecked, "review_required", result.ReviewRequired,
			"failures", result.Failures, "batch_failed", err != nil)
		timer := time.NewTimer(w.interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}
