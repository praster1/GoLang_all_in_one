package main

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestWaitForWorkCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := waitForWork(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v", err)
	}
	ctx, stop := context.WithDeadline(context.Background(), time.Unix(0, 0))
	defer stop()
	if err := waitForWork(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("deadline err=%v", err)
	}
}
