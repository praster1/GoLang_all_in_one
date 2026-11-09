package main

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestCancellationRestoresSignalHandlingAndJoins(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var restored atomic.Bool
	done := restoreOnCancel(ctx, func() { restored.Store(true) })
	select {
	case <-done:
		t.Fatal("restored before cancellation")
	default:
	}
	cancel()
	select {
	case <-done:
		if !restored.Load() {
			t.Fatal("finished without restoring")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("signal restoration did not finish")
	}
}
