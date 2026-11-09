// Package batch runs a bounded number of cooperative workers and preserves input order.
package batch

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
	"unicode/utf8"
)

const (
	MaxItems      = 10000
	MaxItemBytes  = 4096
	MaxWorkers    = 32
	MaxAttempts   = 5
	MaxRetryDelay = time.Second
)

var (
	ErrConfig    = errors.New("invalid batch configuration")
	ErrInput     = errors.New("invalid batch input")
	ErrTemporary = errors.New("temporary failure")
	ErrOutput    = errors.New("invalid processor output")
)

type State string

const (
	Succeeded State = "succeeded"
	Failed    State = "failed"
	Canceled  State = "canceled"
)

type Config struct {
	Workers    int
	Attempts   int
	RetryDelay time.Duration
}

type Job struct {
	Index int
	Input string
}

type Result struct {
	Job      Job
	Output   string
	State    State
	Attempts int
	Err      error
}

// Processor must be safe for concurrent calls, cooperate with ctx, and not panic.
// A job's retries are sequential, and returning nil means that attempt succeeded.
type Processor func(ctx context.Context, job Job) (string, error)

// BatchError reports item failures after all workers have joined.
type BatchError struct {
	Count  int
	causes []error
}

func (e *BatchError) Error() string   { return fmt.Sprintf("%d batch item(s) failed", e.Count) }
func (e *BatchError) Unwrap() []error { return append([]error(nil), e.causes...) }

func ValidateConfig(cfg Config) error {
	if cfg.Workers < 1 || cfg.Workers > MaxWorkers {
		return fmt.Errorf("%w: workers must be 1..%d", ErrConfig, MaxWorkers)
	}
	if cfg.Attempts < 1 || cfg.Attempts > MaxAttempts {
		return fmt.Errorf("%w: attempts must be 1..%d", ErrConfig, MaxAttempts)
	}
	if cfg.RetryDelay < 0 || cfg.RetryDelay > MaxRetryDelay {
		return fmt.Errorf("%w: retry delay must be 0..%s", ErrConfig, MaxRetryDelay)
	}
	return nil
}

// Run returns one result per input after every worker has stopped.
// Invalid arguments return nil results. Cancellation preserves completed results
// and marks unfinished items canceled. The caller must not mutate inputs during Run.
func Run(ctx context.Context, inputs []string, cfg Config, process Processor) ([]Result, error) {
	return run(ctx, inputs, cfg, process, waitRetry)
}

func run(ctx context.Context, inputs []string, cfg Config, process Processor,
	wait func(context.Context, time.Duration) error) ([]Result, error) {
	if ctx == nil || process == nil || wait == nil {
		return nil, fmt.Errorf("%w: context, processor and wait function are required", ErrConfig)
	}
	if err := ValidateConfig(cfg); err != nil {
		return nil, err
	}
	if len(inputs) > MaxItems {
		return nil, fmt.Errorf("%w: at most %d items", ErrInput, MaxItems)
	}
	for i, input := range inputs {
		if len(input) > MaxItemBytes || !utf8.ValidString(input) {
			return nil, fmt.Errorf("%w: item %d must be valid UTF-8 and at most %d bytes", ErrInput, i, MaxItemBytes)
		}
	}
	results := make([]Result, len(inputs))
	for i, input := range inputs {
		results[i].Job = Job{Index: i, Input: input}
	}
	if err := ctx.Err(); err != nil {
		markUnfinished(results, err)
		return results, err
	}
	if len(inputs) == 0 {
		return results, nil
	}
	workers := min(cfg.Workers, len(inputs))
	jobs := make(chan Job)
	var group sync.WaitGroup
	for range workers {
		group.Go(func() {
			for job := range jobs {
				results[job.Index] = execute(ctx, job, cfg, process, wait)
			}
		})
	}
	// This goroutine is the only sender and the only closer of jobs.
dispatch:
	for i, input := range inputs {
		if ctx.Err() != nil {
			break
		}
		select {
		case <-ctx.Done():
			break dispatch
		case jobs <- Job{Index: i, Input: input}:
		}
	}
	close(jobs)
	group.Wait()
	if err := ctx.Err(); err != nil {
		markUnfinished(results, err)
		return results, err
	}
	failures := []error{}
	for _, result := range results {
		if result.State == Failed {
			failures = append(failures, fmt.Errorf("item %d: %w", result.Job.Index, result.Err))
		}
	}
	if len(failures) > 0 {
		return results, &BatchError{Count: len(failures), causes: failures}
	}
	return results, nil
}

func execute(ctx context.Context, job Job, cfg Config, process Processor,
	wait func(context.Context, time.Duration) error) Result {
	result := Result{Job: job}
	for attempt := 1; attempt <= cfg.Attempts; attempt++ {
		if err := ctx.Err(); err != nil {
			result.State, result.Err = Canceled, err
			return result
		}
		result.Attempts = attempt
		value, err := process(ctx, job)
		if err == nil {
			if len(value) > MaxItemBytes || !utf8.ValidString(value) {
				result.State, result.Err = Failed, ErrOutput
				return result
			}
			result.State, result.Output = Succeeded, value
			return result
		}
		if ctxErr := ctx.Err(); ctxErr != nil {
			result.State, result.Err = Canceled, ctxErr
			return result
		}
		if !errors.Is(err, ErrTemporary) || attempt == cfg.Attempts {
			result.State, result.Err = Failed, err
			return result
		}
		if err := wait(ctx, retryDelay(cfg.RetryDelay, attempt)); err != nil {
			result.State, result.Err = Canceled, err
			return result
		}
	}
	// ValidateConfig makes the loop execute at least once.
	return result
}

func markUnfinished(results []Result, err error) {
	for i := range results {
		if results[i].State == "" {
			results[i].State, results[i].Err = Canceled, err
		}
	}
}

func retryDelay(base time.Duration, attempt int) time.Duration {
	delay := base
	for i := 1; i < attempt; i++ {
		if delay >= MaxRetryDelay/2 {
			return MaxRetryDelay
		}
		delay *= 2
	}
	return delay
}

func waitRetry(ctx context.Context, delay time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if delay == 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return ctx.Err()
	}
}
