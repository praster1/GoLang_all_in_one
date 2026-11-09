package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"

	"github.com/praster1/GoLang_all_in_one/examples/30-todo-api/todo"
)

func withDeadline(timeout time.Duration, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type waitingRepository struct{ todo.Repository }

func (waitingRepository) Get(ctx context.Context, id int64) (todo.Task, error) {
	<-ctx.Done()
	return todo.Task{}, ctx.Err()
}
func run() error {
	handler, err := todo.NewHandler(waitingRepository{}, nil)
	if err != nil {
		return err
	}
	limited := withDeadline(time.Millisecond, handler)
	response := httptest.NewRecorder()
	limited.ServeHTTP(response, httptest.NewRequest("GET", "/todos/1", nil))
	fmt.Printf("상태: %d\n본문: %s\n", response.Code, strings.TrimSpace(response.Body.String()))
	return nil
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
