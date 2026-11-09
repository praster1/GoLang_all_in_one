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

func TestRequestStatus(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodHead {
			t.Errorf("method=%s", r.Method)
		}
		calls.Add(1)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	client := server.Client()
	client.Timeout = 2 * time.Second
	code, err := requestStatus(context.Background(), client, server.URL)
	if err != nil || code != 204 {
		t.Fatalf("code=%d err=%v", code, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := requestStatus(ctx, client, server.URL); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled=%v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("calls=%d", calls.Load())
	}
}
func TestRequestStatusInFlightCancellation(t *testing.T) {
	started := make(chan struct{})
	stopped := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { close(started); <-r.Context().Done(); close(stopped) }))
	defer server.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	client := server.Client()
	client.Timeout = 2 * time.Second
	result := make(chan error, 1)
	go func() { _, err := requestStatus(ctx, client, server.URL); result <- err }()
	select {
	case <-started:
		cancel()
	case err := <-result:
		t.Fatalf("request ended before start: %v", err)
	}
	if err := <-result; !errors.Is(err, context.Canceled) {
		t.Fatalf("error=%v", err)
	}
	<-stopped
}
