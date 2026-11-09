package main

import (
	"context"
	"errors"
	"testing"

	"github.com/praster1/GoLang_all_in_one/examples/30-todo-api/todo"
)

func TestEmptyFilteredPageStillAdvances(t *testing.T) {
	ctx := context.Background()
	repo := todo.NewMemory()
	for i := 1; i <= 3; i++ {
		task, err := repo.Create(ctx, "task")
		if err != nil {
			t.Fatal(err)
		}
		if i < 3 {
			if _, err := repo.Replace(ctx, task.ID, task.Title, true); err != nil {
				t.Fatal(err)
			}
		}
	}
	first, err := openPage(ctx, repo, 0, 2)
	if err != nil || len(first.Tasks) != 0 || first.NextAfter != 2 || !first.HasMore {
		t.Fatalf("first=%+v %v", first, err)
	}
	next, err := openPage(ctx, repo, first.NextAfter, 2)
	if err != nil || len(next.Tasks) != 1 || next.Tasks[0].ID != 3 || next.HasMore {
		t.Fatalf("next=%+v %v", next, err)
	}
	empty, err := openPage(ctx, repo, 3, 2)
	if err != nil || empty.Tasks == nil || empty.NextAfter != 3 || empty.HasMore {
		t.Fatalf("empty=%+v %v", empty, err)
	}
	for _, size := range []int{0, 100} {
		if _, err := openPage(ctx, repo, 0, size); err == nil {
			t.Fatal("accepted size")
		}
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := openPage(canceled, repo, 0, 2); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
