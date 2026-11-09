package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetchStatus(t *testing.T) {
	for _, code := range []int{200, 204, 404, 503} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(code) }))
		client := server.Client()
		client.Timeout = 2 * time.Second
		_, err := fetch(context.Background(), client, server.URL)
		server.Close()
		var status *StatusError
		if code == 200 {
			if err != nil {
				t.Fatal(err)
			}
		} else if !errors.As(err, &status) || status.Code != code {
			t.Errorf("code=%d error=%v", code, err)
		}
	}
}
