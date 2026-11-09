package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"
)

func newHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		if _, err := io.WriteString(w, "목록"); err != nil {
			return
		}
	})
	mux.HandleFunc("POST /lessons", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		if _, err := io.WriteString(w, "생성됨"); err != nil {
			return
		}
	})
	mux.HandleFunc("GET /lessons/{id}", func(w http.ResponseWriter, r *http.Request) {
		if _, err := fmt.Fprintf(w, "학습 %s", r.PathValue("id")); err != nil {
			return
		}
	})
	return mux
}

func call(client *http.Client, address, method, body string) (int, string, error) {
	req, err := http.NewRequestWithContext(context.Background(), method, address, strings.NewReader(body))
	if err != nil {
		return 0, "", err
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 4097))
	if err != nil {
		return 0, "", err
	}
	if len(data) > 4096 {
		return 0, "", fmt.Errorf("응답 한도 초과")
	}
	return resp.StatusCode, string(data), nil
}

func run() error {
	server := httptest.NewServer(newHandler())
	defer server.Close()
	client := server.Client()
	client.Timeout = 2 * time.Second
	defer client.CloseIdleConnections()
	for _, tc := range []struct{ method, path, body string }{
		{http.MethodGet, "/", ""},
		{http.MethodPost, "/lessons", ""},
		{http.MethodGet, "/lessons/7", ""},
		{http.MethodDelete, "/lessons/7", ""},
		{http.MethodGet, "/missing", ""},
	} {
		status, text, err := call(client, server.URL+tc.path, tc.method, tc.body)
		if err != nil {
			return err
		}
		fmt.Printf("%s %s -> %d: %s\n", tc.method, tc.path, status, strings.TrimSpace(text))
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
