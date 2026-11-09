package main

import (
	"context"
	"fmt"
	"os"

	"github.com/praster1/GoLang_all_in_one/examples/30-todo-api/todo"
)

func run() error {
	ctx := context.Background()
	repository := todo.NewMemory()
	for _, title := range []string{"첫째", "둘째", "셋째", "넷째", "다섯째"} {
		if _, err := repository.Create(ctx, title); err != nil {
			return err
		}
	}
	var after int64
	for number := 1; ; number++ {
		tasks, err := repository.List(ctx, after, 2)
		if err != nil {
			return err
		}
		if len(tasks) == 0 {
			break
		}
		ids := make([]int64, 0, len(tasks))
		for _, task := range tasks {
			ids = append(ids, task.ID)
		}
		fmt.Printf("페이지 %d: %v\n", number, ids)
		after = tasks[len(tasks)-1].ID
	}
	return nil
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
