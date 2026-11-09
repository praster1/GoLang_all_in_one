package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHealth(t *testing.T) {
	server := httptest.NewServer(newHandler())
	defer server.Close()
	client := server.Client()
	client.Timeout = 2 * time.Second
	for _, tc := range []struct {
		method, path string
		status       int
		body         string
	}{{"GET", "/health", 200, "ready\n"}, {"HEAD", "/health", 200, ""}, {"POST", "/health", 405, "Method Not Allowed\n"}, {"GET", "/missing", 404, "404 page not found\n"}} {
		status, body, err := call(client, server.URL+tc.path, tc.method, "")
		if err != nil || status != tc.status || body != tc.body {
			t.Errorf("%s %s: status=%d body=%q err=%v", tc.method, tc.path, status, body, err)
		}
	}
	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	w := httptest.NewRecorder()
	newHandler().ServeHTTP(w, req)
	if w.Header().Get("Allow") != "GET, HEAD" {
		t.Errorf("Allow=%q", w.Header().Get("Allow"))
	}
}
