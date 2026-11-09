package batch

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func config() Config { return Config{Workers: 3, Attempts: 3} }
func upper(ctx context.Context, job Job) (string, error) {
	return strings.ToUpper(job.Input), ctx.Err()
}
func TestResultsFollowInputOrder(t *testing.T) {
	got, err := Run(context.Background(), []string{"c", "a", "한글", "a"}, config(), upper)
	if err != nil || len(got) != 4 {
		t.Fatal(got, err)
	}
	for i, want := range []string{"C", "A", "한글", "A"} {
		if got[i].Job.Index != i || got[i].Output != want || got[i].State != Succeeded || got[i].Attempts != 1 || got[i].Err != nil {
			t.Fatal(got)
		}
	}
}
func TestInvalidArgumentsDoNotStart(t *testing.T) {
	var calls atomic.Int32
	process := func(context.Context, Job) (string, error) { calls.Add(1); return "", nil }
	for _, cfg := range []Config{{Workers: 0, Attempts: 1}, {Workers: 33, Attempts: 1}, {Workers: 1, Attempts: 0}, {Workers: 1, Attempts: 6}, {Workers: 1, Attempts: 1, RetryDelay: -1}, {Workers: 1, Attempts: 1, RetryDelay: 2 * time.Second}} {
		if got, err := Run(context.Background(), []string{"a"}, cfg, process); got != nil || !errors.Is(err, ErrConfig) {
			t.Fatal(got, err)
		}
	}
	for _, inputs := range [][]string{make([]string, MaxItems+1), {strings.Repeat("x", MaxItemBytes+1)}, {string([]byte{0xff})}} {
		if got, err := Run(context.Background(), inputs, config(), process); got != nil || !errors.Is(err, ErrInput) {
			t.Fatal(len(got), err)
		}
	}
	if got, err := Run(context.Background(), nil, config(), nil); got != nil || !errors.Is(err, ErrConfig) {
		t.Fatal(got, err)
	}
	if got, err := Run(nil, nil, config(), process); got != nil || !errors.Is(err, ErrConfig) {
		t.Fatal(got, err)
	}
	if calls.Load() != 0 {
		t.Fatal(calls.Load())
	}
}
func TestEmptyAndMaximumInput(t *testing.T) {
	got, err := Run(context.Background(), nil, config(), upper)
	if err != nil || got == nil || len(got) != 0 {
		t.Fatal(got, err)
	}
	inputs := make([]string, MaxItems)
	inputs[0] = strings.Repeat("a", MaxItemBytes)
	got, err = Run(context.Background(), inputs, Config{Workers: MaxWorkers, Attempts: 1}, upper)
	if err != nil || len(got) != MaxItems || len(got[0].Output) != MaxItemBytes {
		t.Fatal(len(got), err)
	}
}
func TestOnlyTemporaryFailuresRetry(t *testing.T) {
	calls := 0
	delays := []time.Duration{}
	process := func(context.Context, Job) (string, error) {
		calls++
		if calls < 3 {
			return "", fmt.Errorf("busy: %w", ErrTemporary)
		}
		return "ok", nil
	}
	wait := func(_ context.Context, d time.Duration) error { delays = append(delays, d); return nil }
	got, err := run(context.Background(), []string{"a"}, Config{Workers: 1, Attempts: 3, RetryDelay: 10 * time.Millisecond}, process, wait)
	if err != nil || got[0].Attempts != 3 || got[0].Output != "ok" || calls != 3 || !reflect.DeepEqual(delays, []time.Duration{10 * time.Millisecond, 20 * time.Millisecond}) {
		t.Fatal(got, err, calls, delays)
	}
	permanent := errors.New("invalid data")
	calls = 0
	got, err = Run(context.Background(), []string{"a"}, config(), func(context.Context, Job) (string, error) { calls++; return "", permanent })
	var batchError *BatchError
	if !errors.As(err, &batchError) || batchError.Count != 1 || !errors.Is(err, permanent) || calls != 1 || got[0].Attempts != 1 {
		t.Fatal(got, err, calls)
	}
}
func TestRetryExhaustionAndOutputValidation(t *testing.T) {
	got, err := Run(context.Background(), []string{"a"}, config(), func(context.Context, Job) (string, error) { return "", ErrTemporary })
	if !errors.Is(err, ErrTemporary) || got[0].Attempts != 3 || got[0].State != Failed {
		t.Fatal(got, err)
	}
	for _, output := range []string{strings.Repeat("x", MaxItemBytes+1), string([]byte{0xff})} {
		got, err = Run(context.Background(), []string{"a"}, config(), func(context.Context, Job) (string, error) { return output, nil })
		if !errors.Is(err, ErrOutput) || got[0].Attempts != 1 || got[0].Output != "" {
			t.Fatal(got, err)
		}
	}
}
func TestWorkerBoundAndJoin(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	const workers = 3
	started := make(chan struct{}, workers)
	release := make(chan struct{})
	defer close(release)
	var active, peak atomic.Int32
	finished := make(chan []Result, 1)
	errCh := make(chan error, 1)
	process := func(ctx context.Context, job Job) (string, error) {
		n := active.Add(1)
		defer active.Add(-1)
		for {
			old := peak.Load()
			if old >= n || peak.CompareAndSwap(old, n) {
				break
			}
		}
		if job.Index < workers {
			started <- struct{}{}
		}
		select {
		case <-release:
			return job.Input, nil
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	go func() {
		results, err := Run(ctx, []string{"a", "b", "c", "d", "e", "f"}, Config{Workers: workers, Attempts: 1}, process)
		finished <- results
		errCh <- err
	}()
	for range workers {
		select {
		case <-started:
		case <-time.After(5 * time.Second):
			t.Fatal("workers not started")
		}
	}
	if active.Load() != workers {
		t.Fatal(active.Load())
	}
	cancel()
	select {
	case results := <-finished:
		if len(results) != 6 {
			t.Fatal(results)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("workers not joined")
	}
	if !errors.Is(<-errCh, context.Canceled) || active.Load() != 0 || peak.Load() != workers {
		t.Fatal(active.Load(), peak.Load())
	}
}
func TestPreCanceledAndCancellationDuringBackoff(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	calls := 0
	got, err := Run(ctx, []string{"a", "b"}, config(), func(context.Context, Job) (string, error) { calls++; return "", nil })
	if !errors.Is(err, context.Canceled) || calls != 0 {
		t.Fatal(got, err, calls)
	}
	for _, result := range got {
		if result.State != Canceled || result.Attempts != 0 || !errors.Is(result.Err, context.Canceled) {
			t.Fatal(result)
		}
	}
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	got, err = run(ctx, []string{"a"}, config(), func(context.Context, Job) (string, error) { calls++; return "", ErrTemporary }, func(ctx context.Context, _ time.Duration) error { cancel(); return ctx.Err() })
	if !errors.Is(err, context.Canceled) || calls != 1 || got[0].Attempts != 1 || got[0].State != Canceled {
		t.Fatal(got, err, calls)
	}
}
func TestCancellationPreservesCompletedItems(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	got, err := Run(ctx, []string{"first", "cancel", "last"}, Config{Workers: 1, Attempts: 1}, func(ctx context.Context, job Job) (string, error) {
		if job.Index == 1 {
			cancel()
			return "", ctx.Err()
		}
		return job.Input, nil
	})
	if !errors.Is(err, context.Canceled) || got[0].State != Succeeded || got[0].Output != "first" || got[1].State != Canceled || got[2].State != Canceled || got[2].Attempts != 0 {
		t.Fatal(got, err)
	}
}
func TestRetryDelayIsCapped(t *testing.T) {
	if retryDelay(400*time.Millisecond, 1) != 400*time.Millisecond || retryDelay(400*time.Millisecond, 2) != 800*time.Millisecond || retryDelay(400*time.Millisecond, 3) != time.Second {
		t.Fatal("bad delays")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := waitRetry(ctx, time.Hour); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
