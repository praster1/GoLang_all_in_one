package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/praster1/GoLang_all_in_one/examples/31-concurrent-batch/batch"
)

func TestExpiredJobDoesNotCallProcessor(t *testing.T) {
	called := false
	process, err := withJobTimeout(0, func(context.Context, batch.Job) (string, error) { called = true; return "ok", nil })
	if err != nil {
		t.Fatal(err)
	}
	value, err := process(context.Background(), batch.Job{})
	if value != "" || !errors.Is(err, context.DeadlineExceeded) || called {
		t.Fatal(value, err, called)
	}
}
func TestTimeoutWrapperBoundaries(t *testing.T) {
	fn := func(ctx context.Context, job batch.Job) (string, error) { return job.Input, ctx.Err() }
	for _, duration := range []time.Duration{-1, 2 * time.Minute} {
		if got, err := withJobTimeout(duration, fn); got != nil || err == nil {
			t.Fatal(err)
		}
	}
	if got, err := withJobTimeout(time.Second, nil); got != nil || err == nil {
		t.Fatal(err)
	}
	process, err := withJobTimeout(time.Second, fn)
	if err != nil {
		t.Fatal(err)
	}
	if got, err := process(context.Background(), batch.Job{Input: "ok"}); got != "ok" || err != nil {
		t.Fatal(got, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if got, err := process(ctx, batch.Job{}); got != "" || !errors.Is(err, context.Canceled) {
		t.Fatal(got, err)
	}
	want := errors.New("permanent")
	process, err = withJobTimeout(time.Second, func(context.Context, batch.Job) (string, error) { return "", want })
	if err != nil {
		t.Fatal(err)
	}
	if _, err := process(context.Background(), batch.Job{}); !errors.Is(err, want) {
		t.Fatal(err)
	}
}
