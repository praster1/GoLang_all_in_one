package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestCanceledBeforeRequest(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls.Add(1) }))
	defer server.Close()
	client := server.Client()
	client.Timeout = 2 * time.Second
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := request(ctx, client, server.URL); !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v", err)
	}
	if calls.Load() != 0 {
		t.Fatalf("calls=%d", calls.Load())
	}
	expired, stop := context.WithDeadline(context.Background(), time.Unix(0, 0))
	defer stop()
	if err := request(expired, client, server.URL); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("deadline err=%v", err)
	}
}
