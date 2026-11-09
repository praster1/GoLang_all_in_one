package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestFetchText(t *testing.T) {
	for _, tc := range []struct {
		name    string
		status  int
		body    string
		wantErr bool
	}{
		{"empty", 200, "", false}, {"exact limit", 200, strings.Repeat("a", 1024), false},
		{"too large", 200, strings.Repeat("a", 1025), true}, {"status failure", 503, "busy", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
				if _, err := io.WriteString(w, tc.body); err != nil {
					return
				}
			}))
			defer server.Close()
			client := server.Client()
			client.Timeout = 2 * time.Second
			got, err := fetchText(context.Background(), client, server.URL)
			if (err != nil) != tc.wantErr {
				t.Fatalf("error=%v, wantErr=%t", err, tc.wantErr)
			}
			if err == nil && got != tc.body {
				t.Fatalf("body length=%d, want %d", len(got), len(tc.body))
			}
		})
	}
}

type trackedBody struct {
	io.Reader
	closed bool
}

func (b *trackedBody) Close() error { b.closed = true; return nil }

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestFetchTextClosesBody(t *testing.T) {
	for _, status := range []int{200, 404} {
		body := &trackedBody{Reader: strings.NewReader("ok")}
		client := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: status, Body: body, Header: make(http.Header)}, nil
		})}
		_, err := fetchText(context.Background(), client, "http://example.invalid")
		if status == 200 && err != nil {
			t.Fatal(err)
		}
		if !body.closed {
			t.Errorf("status %d: body not closed", status)
		}
	}
}
