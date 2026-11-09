package main

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestRedirectPolicies(t *testing.T) {
	var finalCalls atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/final", http.StatusFound) })
	mux.HandleFunc("/final", func(w http.ResponseWriter, r *http.Request) { finalCalls.Add(1) })
	server := httptest.NewTLSServer(mux)
	defer server.Close()
	client := server.Client()
	client.Timeout = 2 * time.Second
	code, _, err := inspect(client, server.URL+"/start")
	if err != nil || code != 200 || finalCalls.Load() != 1 {
		t.Fatalf("follow: code=%d calls=%d err=%v", code, finalCalls.Load(), err)
	}
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	code, location, err := inspect(client, server.URL+"/start")
	if err != nil || code != 302 || location != "/final" || finalCalls.Load() != 1 {
		t.Fatalf("stop: code=%d location=%q calls=%d err=%v", code, location, finalCalls.Load(), err)
	}
}
