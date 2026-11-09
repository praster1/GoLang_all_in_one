package main

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestJoinedCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	started := make(chan struct{})
	finished := make(chan struct{})
	result := make(chan error, 1)
	go func() {
		result <- runJoined(ctx, func(ctx context.Context) error { defer close(finished); close(started); <-ctx.Done(); return ctx.Err() })
	}()
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("start timeout")
	}
	cancel()
	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
		select {
		case <-finished:
		default:
			t.Fatal("returned before task finished")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("join timeout")
	}
}
func TestJoinedBoundaries(t *testing.T) {
	want := errors.New("task error")
	if err := runJoined(context.Background(), func(context.Context) error { return want }); !errors.Is(err, want) {
		t.Fatal(err)
	}
	if err := runJoined(context.Background(), func(context.Context) error { return nil }); err != nil {
		t.Fatal(err)
	}
	if err := runJoined(context.Background(), nil); err == nil {
		t.Fatal("nil task")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	called := false
	if err := runJoined(ctx, func(context.Context) error { called = true; return nil }); !errors.Is(err, context.Canceled) || called {
		t.Fatal(err, called)
	}
}

func Example_main() {
	main()
	// Output:
	// 종료까지 기다림: true
}
