package main

import (
	"context"
	"errors"
	"testing"
)

func TestProcess(t *testing.T) {
	for _, stopAfter := range []int{0, 1, 3} {
		ctx, cancel := context.WithCancel(context.Background())
		calls := 0
		completed, err := process(ctx, []string{"a", "b", "c"}, func(string) error {
			calls++
			if stopAfter != 0 && calls == stopAfter {
				cancel()
			}
			return nil
		})
		cancel()
		if stopAfter == 0 && (completed != 3 || err != nil) {
			t.Fatalf("normal = %d, %v", completed, err)
		}
		if stopAfter != 0 && (completed != stopAfter || !errors.Is(err, context.Canceled)) {
			t.Fatalf("stop %d = %d, %v", stopAfter, completed, err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	calls := 0
	completed, err := process(ctx, []string{"a"}, func(string) error { calls++; return nil })
	if completed != 0 || calls != 0 || !errors.Is(err, context.Canceled) {
		t.Fatal(completed, calls, err)
	}
	want := errors.New("step failure")
	completed, err = process(context.Background(), []string{"a"}, func(string) error { return want })
	if completed != 0 || !errors.Is(err, want) {
		t.Fatal(completed, err)
	}
	completed, err = process(context.Background(), nil, func(string) error { t.Fatal("called"); return nil })
	if completed != 0 || err != nil {
		t.Fatal(completed, err)
	}
}

func Example_main() {
	main()
	// Output:
	// 완료 항목: 1
	// 취소 확인: true
}
