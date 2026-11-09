package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSearchEscapesQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method=%s", r.Method)
		}
		if r.Header.Get("Accept-Language") != "ko-KR" {
			t.Errorf("language=%s", r.Header.Get("Accept-Language"))
		}
		if r.URL.Query().Get("keep") != "1" {
			t.Error("existing query lost")
		}
		if _, err := fmt.Fprint(w, r.URL.Query().Get("q")); err != nil {
			return
		}
	}))
	defer server.Close()
	client := server.Client()
	client.Timeout = 2 * time.Second
	for _, term := range []string{"", "Go & HTTP", "한글 + ? #"} {
		got, err := search(context.Background(), client, server.URL+"?keep=1", term)
		if err != nil || got != term {
			t.Errorf("term=%q got=%q err=%v", term, got, err)
		}
	}
}
