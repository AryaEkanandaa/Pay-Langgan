package billing

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"testing/synctest"
	"time"
)

type runnerFunc func(context.Context, time.Time, int) (RenewalResult, error)

func (f runnerFunc) Run(ctx context.Context, at time.Time, limit int) (RenewalResult, error) {
	return f(ctx, at, limit)
}

func TestSchedulerSequentialAndContinuesAfterFailures(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		var logs bytes.Buffer
		var starts []time.Time
		runner := runnerFunc(func(batch context.Context, at time.Time, limit int) (RenewalResult, error) {
			if limit != 42 {
				t.Fatalf("limit = %d", limit)
			}
			starts = append(starts, at)
			switch len(starts) {
			case 1:
				time.Sleep(5 * time.Second)
				return RenewalResult{}, errors.New("temporary database failure")
			case 2:
				return RenewalResult{Failures: []RenewalFailure{{SubscriptionID: 7, Error: "invalid plan"}}}, nil
			case 3:
				return RenewalResult{InvoicesCreated: 2}, nil
			default:
				cancel()
				return RenewalResult{}, batch.Err()
			}
		})
		scheduler, err := NewScheduler(runner, slog.New(slog.NewJSONHandler(&logs, nil)), 2*time.Second, 10*time.Second, 42)
		if err != nil {
			t.Fatal(err)
		}
		scheduler.Run(ctx)
		if len(starts) != 4 || starts[1].Sub(starts[0]) != 7*time.Second || starts[2].Sub(starts[1]) != 2*time.Second {
			t.Fatalf("unexpected scheduling: %v", starts)
		}
		for _, event := range []string{"billing_batch_failed", "billing_batch_partial", "billing_subscription_failed", "billing_batch_completed", "billing_worker_stopped"} {
			if !strings.Contains(logs.String(), event) {
				t.Fatalf("missing log %s", event)
			}
		}
	})
}

func TestSchedulerTimeoutAndShutdown(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		var logs bytes.Buffer
		calls := 0
		runner := runnerFunc(func(batch context.Context, _ time.Time, _ int) (RenewalResult, error) {
			calls++
			if calls == 2 {
				cancel()
			}
			<-batch.Done()
			if calls == 1 && !errors.Is(batch.Err(), context.DeadlineExceeded) {
				t.Fatal("batch deadline missing")
			}
			return RenewalResult{}, batch.Err()
		})
		scheduler, err := NewScheduler(runner, slog.New(slog.NewJSONHandler(&logs, nil)), time.Second, 3*time.Second, 1)
		if err != nil {
			t.Fatal(err)
		}
		start := time.Now()
		scheduler.Run(ctx)
		if calls != 2 || time.Since(start) != 4*time.Second {
			t.Fatal("timeout retry or shutdown failed")
		}
	})
}

func TestSchedulerStopsWhileWaiting(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		calls := 0
		runner := runnerFunc(func(context.Context, time.Time, int) (RenewalResult, error) { calls++; return RenewalResult{}, nil })
		scheduler, err := NewScheduler(runner, slog.Default(), time.Hour, time.Minute, 1)
		if err != nil {
			t.Fatal(err)
		}
		done := make(chan struct{})
		go func() { scheduler.Run(ctx); close(done) }()
		synctest.Wait()
		cancel()
		<-done
		if calls != 1 {
			t.Fatalf("calls = %d", calls)
		}
		scheduler.Run(ctx)
		if calls != 1 {
			t.Fatal("cancelled scheduler performed work")
		}
	})
}
