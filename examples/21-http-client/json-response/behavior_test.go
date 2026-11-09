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

func TestDecodeLesson(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		valid      bool
	}{
		{"valid", `{"title":"Go","minutes":1}`, true}, {"maximum", `{"title":"Go","minutes":180}`, true},
		{"trailing space", "{\"title\":\"Go\",\"minutes\":25}\n ", true},
		{"empty", "", false}, {"null", "null", false}, {"missing", `{"title":"Go"}`, false},
		{"unknown", `{"title":"Go","minutes":25,"extra":1}`, false},
		{"two values", `{"title":"Go","minutes":25}{}`, false}, {"trailing junk", `{"title":"Go","minutes":25}x`, false},
		{"blank title", `{"title":" ","minutes":25}`, false}, {"zero", `{"title":"Go","minutes":0}`, false},
		{"excess", `{"title":"Go","minutes":181}`, false}, {"wrong type", `{"title":"Go","minutes":"25"}`, false},
		{"exact byte limit", `{"title":"Go","minutes":25}` + strings.Repeat(" ", 256-len(`{"title":"Go","minutes":25}`)), true},
		{"one byte over limit", `{"title":"Go","minutes":25}` + strings.Repeat(" ", 257-len(`{"title":"Go","minutes":25}`)), false},
		{"oversize", strings.Repeat(" ", 257), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := decodeLesson(strings.NewReader(tc.body))
			if (err == nil) != tc.valid {
				t.Errorf("error=%v, valid=%t", err, tc.valid)
			}
		})
	}
}
func TestFetchLessonProtocol(t *testing.T) {
	for _, tc := range []struct {
		status int
		media  string
		valid  bool
	}{{200, "application/json", true}, {200, "application/json; charset=utf-8", true}, {200, "text/plain", false}, {200, "", false}, {503, "application/json", false}} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", tc.media)
			w.WriteHeader(tc.status)
			if _, err := io.WriteString(w, `{"title":"Go","minutes":25}`); err != nil {
				return
			}
		}))
		client := server.Client()
		client.Timeout = 2 * time.Second
		_, err := fetchLesson(context.Background(), client, server.URL)
		server.Close()
		if (err == nil) != tc.valid {
			t.Errorf("status=%d media=%q err=%v", tc.status, tc.media, err)
		}
	}
}
