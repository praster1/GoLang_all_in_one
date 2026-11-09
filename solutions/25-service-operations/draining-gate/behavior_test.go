package main

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestAlreadyAcceptedWorkFinishes(t *testing.T) {
	var accepting atomic.Bool
	accepting.Store(true)
	entered, release, done := make(chan struct{}), make(chan struct{}), make(chan struct{})
	handler := gate(&accepting, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { close(entered); <-release; w.WriteHeader(204) }))
	first := httptest.NewRecorder()
	go func() { defer close(done); handler.ServeHTTP(first, httptest.NewRequest("POST", "/work", nil)) }()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("not entered")
	}
	accepting.Store(false)
	second := httptest.NewRecorder()
	handler.ServeHTTP(second, httptest.NewRequest("POST", "/work", nil))
	close(release)
	<-done
	if first.Code != 204 || second.Code != 503 || second.Header().Get("Retry-After") != "1" {
		t.Fatalf("first=%d second=%d", first.Code, second.Code)
	}
}
