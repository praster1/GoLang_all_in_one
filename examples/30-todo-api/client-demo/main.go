package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"

	"github.com/praster1/GoLang_all_in_one/examples/30-todo-api/todo"
)

func run() error {
	handler, err := todo.NewHandler(todo.NewMemory(), nil)
	if err != nil {
		return err
	}
	server := httptest.NewServer(handler)
	defer server.Close()
	client := server.Client()
	client.Timeout = time.Second
	defer client.CloseIdleConnections()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, "POST", server.URL+"/todos", strings.NewReader(`{"title":"HTTP 실습"}`))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	body, readErr := io.ReadAll(io.LimitReader(response.Body, 4097))
	closeErr := response.Body.Close()
	if err := errors.Join(readErr, closeErr); err != nil {
		return err
	}
	if len(body) > 4096 {
		return errors.New("응답 크기 초과")
	}
	if response.StatusCode != 201 {
		return fmt.Errorf("예상과 다른 상태: %d", response.StatusCode)
	}
	var task todo.Task
	if err := json.Unmarshal(body, &task); err != nil {
		return err
	}
	fmt.Printf("생성: %d %s 완료=%t\n", task.ID, task.Title, task.Done)
	fmt.Println("위치:", response.Header.Get("Location"))
	return nil
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
