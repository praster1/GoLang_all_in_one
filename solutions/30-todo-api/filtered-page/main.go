package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/praster1/GoLang_all_in_one/examples/30-todo-api/todo"
)

type OpenPage struct {
	Tasks     []todo.Task
	NextAfter int64
	HasMore   bool
}

func openPage(ctx context.Context, repo todo.Repository, after int64, size int) (OpenPage, error) {
	if size < 1 || size > 99 {
		return OpenPage{}, errors.New("size는 1~99여야 합니다")
	}
	source, err := repo.List(ctx, after, size+1)
	if err != nil {
		return OpenPage{}, err
	}
	page := OpenPage{Tasks: make([]todo.Task, 0), NextAfter: after, HasMore: len(source) > size}
	if page.HasMore {
		source = source[:size]
	}
	for _, task := range source {
		page.NextAfter = task.ID
		if !task.Done {
			page.Tasks = append(page.Tasks, task)
		}
	}
	return page, nil
}
func run() error {
	ctx := context.Background()
	repo := todo.NewMemory()
	for i := 1; i <= 5; i++ {
		task, err := repo.Create(ctx, fmt.Sprintf("항목 %d", i))
		if err != nil {
			return err
		}
		if i%2 == 1 {
			if _, err := repo.Replace(ctx, task.ID, task.Title, true); err != nil {
				return err
			}
		}
	}
	var after int64
	for number := 1; ; number++ {
		page, err := openPage(ctx, repo, after, 2)
		if err != nil {
			return err
		}
		ids := make([]int64, 0, len(page.Tasks))
		for _, task := range page.Tasks {
			ids = append(ids, task.ID)
		}
		fmt.Printf("페이지 %d: 미완료=%v 다음=%d 더있음=%t\n", number, ids, page.NextAfter, page.HasMore)
		if !page.HasMore {
			break
		}
		after = page.NextAfter
	}
	return nil
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
