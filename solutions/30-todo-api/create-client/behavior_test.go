package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/praster1/GoLang_all_in_one/examples/30-todo-api/todo"
)

func TestCreateClient(t *testing.T) {
	handler, err := todo.NewHandler(todo.NewMemory(), nil)
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(handler)
	defer server.Close()
	client := server.Client()
	client.Timeout = time.Second
	defer client.CloseIdleConnections()
	task, err := createTask(context.Background(), client, server.URL, " Go ")
	if err != nil || task.Title != "Go" || task.ID != 1 {
		t.Fatalf("task=%+v %v", task, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := createTask(ctx, client, server.URL, "next"); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
func TestClientFailureContracts(t *testing.T) {
	cases := []struct {
		status    int
		body      string
		apiStatus int
	}{{500, `{"error":"internal_error"}`, 500}, {400, "broken", 400}, {201, "bad", 0}, {201, `{"id":0,"title":"Go","done":false}`, 0}, {201, strings.Repeat(" ", 4097), 0}}
	for _, tc := range cases {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(tc.status)
			_, _ = io.WriteString(w, tc.body)
		}))
		client := server.Client()
		client.Timeout = time.Second
		_, err := createTask(context.Background(), client, server.URL, "Go")
		client.CloseIdleConnections()
		server.Close()
		if err == nil {
			t.Fatal("accepted invalid response")
		}
		if tc.apiStatus != 0 {
			var api *APIError
			if !errors.As(err, &api) || api.Status != tc.apiStatus {
				t.Fatalf("error=%v", err)
			}
		}
	}
	if _, err := createTask(context.Background(), &http.Client{}, "ftp://example.invalid", "Go"); err == nil {
		t.Fatal("accepted URL")
	}
}
