package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

func TestRequestIDReplacesUntrustedHeader(t *testing.T) {
	var calls atomic.Int32
	handler := withRequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		id := requestID(r.Context())
		if id == "" || id != w.Header().Get("X-Request-ID") {
			t.Errorf("context id=%q header=%q", id, w.Header().Get("X-Request-ID"))
		}
		w.WriteHeader(204)
	}))
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Request-ID", "client-supplied")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if calls.Load() != 1 || w.Code != 204 || w.Header().Get("X-Request-ID") != "r-1" {
		t.Fatalf("calls=%d code=%d id=%q", calls.Load(), w.Code, w.Header().Get("X-Request-ID"))
	}
}
func TestRequestIDConcurrent(t *testing.T) {
	const count = 50
	handler := newHandler()
	ids := make(chan string, count)
	var wg sync.WaitGroup
	for range count {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, httptest.NewRequest("GET", "/ready", nil))
			if w.Code != 200 {
				t.Errorf("status=%d", w.Code)
			}
			ids <- w.Header().Get("X-Request-ID")
		}()
	}
	wg.Wait()
	close(ids)
	seen := make(map[string]bool)
	for id := range ids {
		if !strings.HasPrefix(id, "r-") || seen[id] {
			t.Errorf("invalid or duplicate id=%q", id)
		}
		seen[id] = true
	}
	if len(seen) != count {
		t.Errorf("unique=%d", len(seen))
	}
}
