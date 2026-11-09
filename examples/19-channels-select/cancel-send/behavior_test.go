package main

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestSendCancellationUnblocks(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	out := make(chan int)
	done := make(chan error, 1)
	go func() { done <- sendNumbers(ctx, 1000, out) }()
	select {
	case value := <-out:
		if value != 1 {
			t.Fatal(value)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("first value timeout")
	}
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("cancel did not unblock sender")
	}
	if _, ok := <-out; ok {
		t.Fatal("sender did not close output")
	}
}
func TestSendBoundsAndEmpty(t *testing.T) {
	for _, count := range []int{-1, 1001} {
		out := make(chan int)
		if err := sendNumbers(context.Background(), count, out); err == nil {
			t.Fatal("invalid count accepted")
		}
		if _, ok := <-out; ok {
			t.Fatal("output open")
		}
	}
	out := make(chan int)
	if err := sendNumbers(context.Background(), 0, out); err != nil {
		t.Fatal(err)
	}
	if _, ok := <-out; ok {
		t.Fatal("empty output open")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	out = make(chan int)
	if err := sendNumbers(ctx, 1, out); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
