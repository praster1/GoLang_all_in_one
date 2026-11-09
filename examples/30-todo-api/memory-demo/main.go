package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/praster1/GoLang_all_in_one/examples/30-todo-api/todo"
)

func run() error {
	ctx := context.Background()
	repository := todo.NewMemory()
	first, err := repository.Create(ctx, "Go 복습")
	if err != nil {
		return err
	}
	second, err := repository.Create(ctx, "SQL 복습")
	if err != nil {
		return err
	}
	first, err = repository.Replace(ctx, first.ID, first.Title, true)
	if err != nil {
		return err
	}
	tasks, err := repository.List(ctx, 0, 20)
	if err != nil {
		return err
	}
	for _, task := range tasks {
		fmt.Printf("%d %s 완료=%t\n", task.ID, task.Title, task.Done)
	}
	if err := repository.Delete(ctx, second.ID); err != nil {
		return err
	}
	_, err = repository.Get(ctx, second.ID)
	fmt.Println("삭제 확인:", errors.Is(err, todo.ErrNotFound))
	return nil
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
