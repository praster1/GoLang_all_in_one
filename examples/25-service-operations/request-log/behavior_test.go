package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRequestLog(t *testing.T) {
	cases := []struct {
		name         string
		handle       http.HandlerFunc
		status, size int
	}{
		{"implicit", func(w http.ResponseWriter, r *http.Request) { _, _ = io.WriteString(w, "hello") }, 200, 5},
		{"explicit", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(201)
			w.WriteHeader(500)
			_, _ = io.WriteString(w, "x")
		}, 201, 1},
		{"empty", func(w http.ResponseWriter, r *http.Request) {}, 200, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var output bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&output, nil))
			mux := http.NewServeMux()
			mux.HandleFunc("GET /items/{id}", tc.handle)
			start := time.Unix(0, 0)
			calls := 0
			handler := requestLog(logger, func() time.Time { calls++; return start.Add(time.Duration(calls) * 3 * time.Millisecond) }, mux)
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest("GET", "/items/secret?token=secret", nil))
			var item map[string]any
			if err := json.Unmarshal(output.Bytes(), &item); err != nil {
				t.Fatal(err)
			}
			if item["status"] != float64(tc.status) || item["bytes"] != float64(tc.size) || item["elapsed_ms"] != float64(3) || item["route"] != "GET /items/{id}" || recorder.Code != tc.status {
				t.Fatalf("record=%v HTTP=%d", item, recorder.Code)
			}
			if strings.Contains(output.String(), "secret") {
				t.Fatal("URL content leaked")
			}
		})
	}
}
func TestUnwrap(t *testing.T) {
	r := httptest.NewRecorder()
	w := &responseCapture{ResponseWriter: r}
	if w.Unwrap() != r {
		t.Fatal("lost writer")
	}
}
