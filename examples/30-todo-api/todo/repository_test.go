package todo

import (
	"context"
	"errors"
	"math"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func repositories(t *testing.T) map[string]func(*testing.T) Repository {
	t.Helper()
	return map[string]func(*testing.T) Repository{
		"memory": func(t *testing.T) Repository { return NewMemory() },
		"sqlite": func(t *testing.T) Repository {
			db, err := OpenSQLite(context.Background(), ":memory:")
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := db.Close(); err != nil {
					t.Error(err)
				}
			})
			return db
		},
	}
}
func TestRepositoryContract(t *testing.T) {
	for name, factory := range repositories(t) {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			repo := factory(t)
			empty, err := repo.List(ctx, 0, 20)
			if err != nil || empty == nil || len(empty) != 0 {
				t.Fatalf("empty=%v err=%v", empty, err)
			}
			first, err := repo.Create(ctx, "  Go 복습  ")
			if err != nil || first.ID <= 0 || first.Title != "Go 복습" || first.Done {
				t.Fatalf("create=%+v err=%v", first, err)
			}
			original := first
			first.Title = "caller change"
			found, err := repo.Get(ctx, original.ID)
			if err != nil || found != original {
				t.Fatalf("get=%+v err=%v", found, err)
			}
			second, err := repo.Create(ctx, "SQL 복습")
			if err != nil || second.ID <= original.ID {
				t.Fatalf("second=%+v %v", second, err)
			}
			page, err := repo.List(ctx, original.ID, 1)
			if err != nil || !reflect.DeepEqual(page, []Task{second}) {
				t.Fatalf("page=%v %v", page, err)
			}
			page[0].Title = "caller list change"
			again, _ := repo.Get(ctx, second.ID)
			if again.Title != "SQL 복습" {
				t.Fatal("list aliases storage")
			}
			replaced, err := repo.Replace(ctx, original.ID, " Go 완료 ", true)
			if err != nil || replaced.Title != "Go 완료" || !replaced.Done {
				t.Fatalf("replace=%+v %v", replaced, err)
			}
			replaced, err = repo.Replace(ctx, original.ID, "Go 완료", false)
			if err != nil || replaced.Done {
				t.Fatal("false cannot be saved")
			}
			if _, err := repo.Replace(ctx, original.ID, " ", true); err == nil {
				t.Fatal("accepted empty title")
			}
			afterInvalid, _ := repo.Get(ctx, original.ID)
			if afterInvalid != replaced {
				t.Fatal("invalid replace changed state")
			}
			if err := repo.Delete(ctx, second.ID); err != nil {
				t.Fatal(err)
			}
			if _, err := repo.Get(ctx, second.ID); !errors.Is(err, ErrNotFound) {
				t.Fatalf("deleted get=%v", err)
			}
			if err := repo.Delete(ctx, second.ID); !errors.Is(err, ErrNotFound) {
				t.Fatalf("deleted delete=%v", err)
			}
			third, err := repo.Create(ctx, "새 할 일")
			if err != nil || third.ID <= second.ID {
				t.Fatalf("reused id=%d err=%v", third.ID, err)
			}
			if _, err := repo.Replace(ctx, second.ID, "없는 항목", true); !errors.Is(err, ErrNotFound) {
				t.Fatalf("missing replace=%v", err)
			}
			for _, id := range []int64{0, -1} {
				if _, err := repo.Get(ctx, id); !errors.Is(err, ErrInvalidID) {
					t.Fatal(err)
				}
				if err := repo.Delete(ctx, id); !errors.Is(err, ErrInvalidID) {
					t.Fatal(err)
				}
			}
			for _, p := range []struct {
				after int64
				limit int
			}{{-1, 20}, {0, 0}, {0, 101}} {
				if _, err := repo.List(ctx, p.after, p.limit); !errors.Is(err, ErrInvalidPage) {
					t.Fatalf("page error=%v", err)
				}
			}
		})
	}
}
func TestTitleBoundaries(t *testing.T) {
	for name, factory := range repositories(t) {
		t.Run(name, func(t *testing.T) {
			repo := factory(t)
			ctx := context.Background()
			for _, title := range []string{"", " \n\t", strings.Repeat("가", 81), "middle\nline", string([]byte{0xff})} {
				var validation *ValidationError
				if _, err := repo.Create(ctx, title); !errors.As(err, &validation) {
					t.Fatalf("title=%q error=%v", title, err)
				}
			}
			for _, title := range []string{"가", strings.Repeat("가", 80), "정상 � 문자"} {
				if _, err := repo.Create(ctx, title); err != nil {
					t.Fatal(err)
				}
			}
			tasks, err := repo.List(ctx, 0, 100)
			if err != nil || len(tasks) != 3 {
				t.Fatalf("tasks=%v %v", tasks, err)
			}
		})
	}
}
func TestCanceledRepositoryCallsDoNotChangeState(t *testing.T) {
	for name, factory := range repositories(t) {
		t.Run(name, func(t *testing.T) {
			repo := factory(t)
			background := context.Background()
			original, err := repo.Create(background, "유지")
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(background)
			cancel()
			_, a := repo.Create(ctx, "추가")
			_, b := repo.Get(ctx, original.ID)
			_, c := repo.List(ctx, 0, 20)
			_, d := repo.Replace(ctx, original.ID, "변경", true)
			e := repo.Delete(ctx, original.ID)
			for _, err := range []error{a, b, c, d, e} {
				if !errors.Is(err, context.Canceled) {
					t.Fatalf("error=%v", err)
				}
			}
			tasks, err := repo.List(background, 0, 20)
			if err != nil || !reflect.DeepEqual(tasks, []Task{original}) {
				t.Fatalf("state=%v %v", tasks, err)
			}
		})
	}
}
func TestConcurrentCreate(t *testing.T) {
	for name, factory := range repositories(t) {
		t.Run(name, func(t *testing.T) {
			repo := factory(t)
			const count = 30
			var wg sync.WaitGroup
			results := make(chan Task, count)
			failures := make(chan error, count)
			for range count {
				wg.Add(1)
				go func() {
					defer wg.Done()
					task, err := repo.Create(context.Background(), "동시 생성")
					if err != nil {
						failures <- err
						return
					}
					results <- task
				}()
			}
			wg.Wait()
			close(results)
			close(failures)
			for err := range failures {
				t.Error(err)
			}
			seen := make(map[int64]bool)
			for task := range results {
				if seen[task.ID] {
					t.Fatal("duplicate id")
				}
				seen[task.ID] = true
			}
			tasks, err := repo.List(context.Background(), 0, 100)
			if err != nil || len(tasks) != count || len(seen) != count {
				t.Fatalf("count=%d ids=%d err=%v", len(tasks), len(seen), err)
			}
			for i := 1; i < len(tasks); i++ {
				if tasks[i-1].ID >= tasks[i].ID {
					t.Fatal("not sorted")
				}
			}
		})
	}
}
func TestSQLitePersistence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "todos.db")
	ctx := context.Background()
	first, err := OpenSQLite(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	task, createErr := first.Create(ctx, "저장")
	closeErr := first.Close()
	if err := errors.Join(createErr, closeErr); err != nil {
		t.Fatal(err)
	}
	second, err := OpenSQLite(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	got, err := second.Get(ctx, task.ID)
	if err != nil || got != task {
		t.Fatalf("got=%+v %v", got, err)
	}
}
func TestSQLiteParameterBinding(t *testing.T) {
	db, err := OpenSQLite(context.Background(), ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	title := `x'); DROP TABLE todos; --`
	task, err := db.Create(context.Background(), title)
	if err != nil {
		t.Fatal(err)
	}
	got, err := db.Get(context.Background(), task.ID)
	if err != nil || got.Title != title {
		t.Fatalf("got=%+v %v", got, err)
	}
}
func TestMemoryIDOverflow(t *testing.T) {
	memory := NewMemory()
	memory.nextID = math.MaxInt64
	task, err := memory.Create(context.Background(), "last")
	if err != nil || task.ID != math.MaxInt64 {
		t.Fatalf("task=%+v %v", task, err)
	}
	if _, err := memory.Create(context.Background(), "overflow"); !errors.Is(err, ErrIDExhausted) {
		t.Fatalf("error=%v", err)
	}
}
