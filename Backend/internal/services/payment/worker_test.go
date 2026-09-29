package payment

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"
)

type timeoutMaintenance struct{ calls chan struct{} }

func (r *timeoutMaintenance) RunMaintenance(ctx context.Context, _ int) (MaintenanceResult, error) {
	r.calls <- struct{}{}
	<-ctx.Done()
	return MaintenanceResult{}, ctx.Err()
}

func TestPaymentWorkerTimeoutAndShutdown(t *testing.T) {
	runner := &timeoutMaintenance{calls: make(chan struct{}, 4)}
	worker, err := NewWorker(runner, slog.New(slog.NewTextHandler(io.Discard, nil)), time.Millisecond, 5*time.Millisecond, 10)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() { worker.Run(ctx); close(done) }()
	for i := 0; i < 2; i++ {
		select {
		case <-runner.calls:
		case <-time.After(2 * time.Second):
			t.Fatal("worker did not retry after batch timeout")
		}
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("worker did not stop")
	}
}

func TestPaymentRetryBackoff(t *testing.T) {
	if retryDelay(1) != 30*time.Second || retryDelay(2) != time.Minute || retryDelay(100) != time.Hour {
		t.Fatal("invalid retry backoff")
	}
}
