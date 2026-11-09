package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"time"

	"github.com/praster1/GoLang_all_in_one/examples/30-todo-api/todo"
)

type APIError struct {
	Status int
	Code   string
}

func (e *APIError) Error() string { return fmt.Sprintf("API 상태 %d: %s", e.Status, e.Code) }
func createTask(ctx context.Context, client *http.Client, base, title string) (todo.Task, error) {
	title, err := todo.NormalizeTitle(title)
	if err != nil {
		return todo.Task{}, err
	}
	target, err := url.Parse(base)
	if err != nil || target.Host == "" || (target.Scheme != "http" && target.Scheme != "https") || target.User != nil || target.RawQuery != "" || target.Fragment != "" || (target.Path != "" && target.Path != "/") {
		return todo.Task{}, errors.New("서버 원점 URL이 필요합니다")
	}
	target.Path = "/todos"
	body, err := json.Marshal(struct {
		Title string `json:"title"`
	}{title})
	if err != nil {
		return todo.Task{}, err
	}
	request, err := http.NewRequestWithContext(ctx, "POST", target.String(), bytes.NewReader(body))
	if err != nil {
		return todo.Task{}, err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		return todo.Task{}, err
	}
	data, readErr := io.ReadAll(io.LimitReader(response.Body, 4097))
	closeErr := response.Body.Close()
	if err := errors.Join(readErr, closeErr); err != nil {
		return todo.Task{}, err
	}
	if len(data) > 4096 {
		return todo.Task{}, errors.New("응답 크기 초과")
	}
	if response.StatusCode != 201 {
		var p struct {
			Error string `json:"error"`
		}
		if err := json.Unmarshal(data, &p); err != nil || p.Error == "" {
			return todo.Task{}, &APIError{Status: response.StatusCode, Code: "invalid_error_response"}
		}
		return todo.Task{}, &APIError{Status: response.StatusCode, Code: p.Error}
	}
	var task todo.Task
	if err := json.Unmarshal(data, &task); err != nil {
		return todo.Task{}, err
	}
	if task.ID <= 0 || task.Title != title || task.Done {
		return todo.Task{}, errors.New("생성 응답 계약 오류")
	}
	return task, nil
}
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
	task, err := createTask(ctx, client, server.URL, " Go 복습 ")
	if err != nil {
		return err
	}
	fmt.Printf("생성: %d %s 완료=%t\n", task.ID, task.Title, task.Done)
	return nil
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
