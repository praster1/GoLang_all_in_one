package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"time"
)

type Problem struct {
	Error string `json:"error"`
}

func writeProblem(w http.ResponseWriter, status int, message string) {
	data, err := json.Marshal(Problem{Error: message})
	if err != nil {
		http.Error(w, "응답 생성 실패", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(append(data, '\n')); err != nil {
		return
	}
}

func newHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /lessons/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil || id < 1 {
			writeProblem(w, http.StatusBadRequest, "invalid_id")
			return
		}
		if id != 1 {
			writeProblem(w, http.StatusNotFound, "lesson_not_found")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if _, err := io.WriteString(w, "{\"id\":1,\"title\":\"Go\"}\n"); err != nil {
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
		{http.MethodGet, "/lessons/x", ""},
		{http.MethodGet, "/lessons/2", ""},
		{http.MethodGet, "/lessons/1", ""},
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
