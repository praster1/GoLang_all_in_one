package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthContract(t *testing.T) {
	handler := newHandler()
	for _, tc := range []struct {
		method, path string
		want         int
	}{{"GET", "/healthz", 200}, {"POST", "/healthz", 405}, {"GET", "/missing", 404}} {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest(tc.method, tc.path, nil))
		if w.Code != tc.want {
			t.Fatalf("%s %s: %d", tc.method, tc.path, w.Code)
		}
		if tc.want == http.StatusOK {
			var body struct {
				Status string `json:"status"`
			}
			if w.Header().Get("Content-Type") != "application/json" {
				t.Fatal("media type")
			}
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil || body.Status != "ok" {
				t.Fatal("body contract")
			}
		}
	}
}

func Example_main() {
	main()
	// Output:
	// 200 {"status":"ok"}
}
