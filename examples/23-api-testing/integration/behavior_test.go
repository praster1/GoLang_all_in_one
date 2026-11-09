package main

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPRoundTrip(t *testing.T) {
	server := httptest.NewTLSServer(newHandler(memoryStore{lessons: map[int]Lesson{1: {ID: 1, Title: "Go"}}}))
	defer server.Close()
	client := server.Client()
	client.Timeout = 2 * time.Second
	for _, tc := range []struct {
		method, path string
		status       int
	}{{"GET", "/lessons/1", 200}, {"HEAD", "/lessons/1", 200}, {"GET", "/lessons/2", 404}, {"POST", "/lessons/1", 405}} {
		code, body, err := inspect(client, server.URL+tc.path, tc.method)
		if err != nil || code != tc.status {
			t.Fatalf("%s %s: code=%d err=%v", tc.method, tc.path, code, err)
		}
		if tc.method == "HEAD" {
			if body != "" {
				t.Errorf("HEAD body=%q", body)
			}
		} else if code == 200 {
			var lesson Lesson
			if err := json.Unmarshal([]byte(body), &lesson); err != nil || lesson.ID != 1 {
				t.Errorf("lesson=%+v err=%v", lesson, err)
			}
		}
	}
}
