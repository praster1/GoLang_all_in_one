package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/praster1/GoLang_all_in_one/examples/30-todo-api/todo"
)

func run() (err error) {
	dir, err := os.MkdirTemp("", "todo-book-")
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, os.RemoveAll(dir)) }()
	ctx := context.Background()
	path := filepath.Join(dir, "tasks.db")
	first, err := todo.OpenSQLite(ctx, path)
	if err != nil {
		return err
	}
	task, createErr := first.Create(ctx, "저장한 할 일")
	if err := errors.Join(createErr, first.Close()); err != nil {
		return err
	}
	second, err := todo.OpenSQLite(ctx, path)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, second.Close()) }()
	found, err := second.Get(ctx, task.ID)
	if err != nil {
		return err
	}
	fmt.Printf("다시 열기: %d %s 완료=%t\n", found.ID, found.Title, found.Done)
	return nil
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
