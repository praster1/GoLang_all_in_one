package main

import (
	"net/http/httptest"
	"testing"
)

func TestRouting(t *testing.T) {
	handler := newHandler()
	for _, tc := range []struct {
		method, path string
		status       int
		allow        string
	}{{"GET", "/", 200, ""}, {"POST", "/lessons", 201, ""}, {"GET", "/lessons/7", 200, ""}, {"DELETE", "/lessons/7", 405, "GET, HEAD"}, {"GET", "/missing", 404, ""}, {"GET", "/lessons", 405, "POST"}, {"POST", "/", 405, "GET, HEAD"}} {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest(tc.method, tc.path, nil))
		if w.Code != tc.status || w.Header().Get("Allow") != tc.allow {
			t.Errorf("%s %s: code=%d allow=%q", tc.method, tc.path, w.Code, w.Header().Get("Allow"))
		}
	}
}
