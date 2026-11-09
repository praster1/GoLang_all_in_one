package main

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"
)

func TestHealthPhases(t *testing.T) {
	calls := 0
	var checkErr error
	h := &Health{Check: func(ctx context.Context) error {
		calls++
		if _, ok := ctx.Deadline(); !ok {
			t.Fatal("no deadline")
		}
		return checkErr
	}}
	handler := h.Handler()
	status := func(path string) int {
		r := httptest.NewRecorder()
		handler.ServeHTTP(r, httptest.NewRequest("GET", path, nil))
		return r.Code
	}
	if status("/livez") != 200 || status("/readyz") != 503 || calls != 0 {
		t.Fatal("wrong startup state")
	}
	h.Ready.Store(true)
	if status("/readyz") != 200 || calls != 1 {
		t.Fatal("wrong ready state")
	}
	checkErr = errors.New("private DB information")
	if status("/readyz") != 503 || status("/livez") != 200 {
		t.Fatal("dependency changes liveness")
	}
	h.Ready.Store(false)
	before := calls
	if status("/readyz") != 503 || calls != before {
		t.Fatal("draining called dependency")
	}
}
func TestReadinessCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	h := &Health{Check: func(ctx context.Context) error { return ctx.Err() }}
	h.Ready.Store(true)
	r := httptest.NewRecorder()
	h.Handler().ServeHTTP(r, httptest.NewRequest("GET", "/readyz", nil).WithContext(ctx))
	if r.Code != 503 {
		t.Fatalf("status=%d", r.Code)
	}
}
