package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInvalidConfiguration(t *testing.T) {
	cases := [][]string{{"-addr=0.0.0.0:8080"}, {"-addr=bad"}, {"-store=remote"}, {"-store=sqlite"}, {"-db=unused.db"}, {"extra"}, {"-unknown"}}
	for _, args := range cases {
		if err := run(context.Background(), args, io.Discard); err == nil {
			t.Fatalf("accepted %v", args)
		}
	}
}
func TestCanceledStartup(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := run(ctx, []string{"-addr=127.0.0.1:0"}, io.Discard); !errors.Is(err, context.Canceled) {
		t.Fatalf("error=%v", err)
	}
}

func TestRequestDeadline(t *testing.T) {
	handler := requestDeadline(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := r.Context().Deadline(); !ok {
			t.Error("missing request deadline")
		}
		w.WriteHeader(204)
	}))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
	if w.Code != 204 {
		t.Fatalf("status=%d", w.Code)
	}
}
