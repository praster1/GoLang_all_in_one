package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/praster1/GoLang_all_in_one/examples/30-todo-api/todo"
)

func TestDeadlinePropagation(t *testing.T) {
	called := false
	handler := withDeadline(time.Second, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if _, ok := r.Context().Deadline(); !ok {
			t.Error("missing deadline")
		}
		w.WriteHeader(204)
	}))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest("GET", "/", nil))
	if !called || response.Code != 204 {
		t.Fatal("handler not called")
	}
}
func TestCanceledAndExpiredMapping(t *testing.T) {
	handler, err := todo.NewHandler(waitingRepository{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		parentCanceled bool
		want           int
	}{{false, 504}, {true, 503}} {
		ctx, cancel := context.WithCancel(context.Background())
		if tc.parentCanceled {
			cancel()
		}
		response := httptest.NewRecorder()
		withDeadline(time.Millisecond, handler).ServeHTTP(response, httptest.NewRequest("GET", "/todos/1", nil).WithContext(ctx))
		cancel()
		if response.Code != tc.want {
			t.Fatalf("status=%d want=%d", response.Code, tc.want)
		}
	}
}
