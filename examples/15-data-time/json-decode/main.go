package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

type Task struct {
	Title   string `json:"title"`
	Minutes int    `json:"minutes"`
}

func decodeTask(source io.Reader) (Task, error) {
	decoder := json.NewDecoder(source)
	decoder.DisallowUnknownFields()
	var task Task
	if err := decoder.Decode(&task); err != nil {
		return Task{}, fmt.Errorf("작업 JSON: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err != nil {
			return Task{}, fmt.Errorf("JSON 끝 확인: %w", err)
		}
		return Task{}, errors.New("JSON 값은 하나만 허용합니다")
	}
	task.Title = strings.TrimSpace(task.Title)
	if task.Title == "" || task.Minutes < 1 || task.Minutes > 120 {
		return Task{}, errors.New("제목과 1~120분이 필요합니다")
	}
	return task, nil
}

func run() error {
	task, err := decodeTask(strings.NewReader(`{"title":"Go","minutes":25}`))
	if err != nil {
		return err
	}
	fmt.Printf("%s: %d분\n", task.Title, task.Minutes)
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
