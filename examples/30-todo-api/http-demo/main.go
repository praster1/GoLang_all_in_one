package main

import (
	"fmt"
	"net/http/httptest"
	"os"
	"strings"

	"github.com/praster1/GoLang_all_in_one/examples/30-todo-api/todo"
)

func run() error {
	handler, err := todo.NewHandler(todo.NewMemory(), nil)
	if err != nil {
		return err
	}
	cases := []struct{ method, path, body string }{
		{"POST", "/todos", `{"title":" Go 복습 "}`},
		{"GET", "/todos/1", ""},
		{"PUT", "/todos/1", `{"title":"Go 복습","done":true}`},
		{"GET", "/todos?limit=2", ""},
		{"DELETE", "/todos/1", ""},
		{"GET", "/todos/1", ""},
	}
	for _, tc := range cases {
		request := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		if tc.body != "" {
			request.Header.Set("Content-Type", "application/json")
		}
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		line := fmt.Sprintf("%s %s -> %d", tc.method, tc.path, response.Code)
		if body := strings.TrimSpace(response.Body.String()); body != "" {
			line += " " + body
		}
		fmt.Println(line)
	}
	return nil
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
