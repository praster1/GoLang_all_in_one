package main

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestBudgetCleanup(t *testing.T) {
	parent, cancel := context.WithCancel(context.Background())
	defer cancel()
	var observed context.Context
	if err := withBudget(parent, time.Hour, func(ctx context.Context) error { observed = ctx; return nil }); err != nil {
		t.Fatal(err)
	}
	if observed == nil || !errors.Is(observed.Err(), context.Canceled) || parent.Err() != nil {
		t.Fatal("child cleanup or parent lifetime")
	}
	want := errors.New("task failed")
	if err := withBudget(parent, time.Hour, func(ctx context.Context) error { observed = ctx; return want }); !errors.Is(err, want) {
		t.Fatal(err)
	}
	if !errors.Is(observed.Err(), context.Canceled) {
		t.Fatal("error path leaked child")
	}
}

func TestBudgetNoWork(t *testing.T) {
	called := false
	task := func(context.Context) error { called = true; return nil }
	if err := withBudget(context.Background(), 0, task); !errors.Is(err, context.DeadlineExceeded) || called {
		t.Fatal(err, called)
	}
	if err := withBudget(context.Background(), -time.Second, task); err == nil || called {
		t.Fatal(err, called)
	}
	if err := withBudget(context.Background(), time.Hour, nil); err == nil {
		t.Fatal("nil task")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := withBudget(ctx, time.Hour, task); !errors.Is(err, context.Canceled) || called {
		t.Fatal(err, called)
	}
}

func TestBudgetPropagatesCancellation(t *testing.T) {
	parent, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := withBudget(parent, time.Hour, func(ctx context.Context) error { cancel(); return nil })
	if !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func Example_main() {
	main()
	// Output:
	// 작업 성공: true
	// 반환 후 정리: true
	// 0초 마감: true
	// 0초 작업 실행: false
}
