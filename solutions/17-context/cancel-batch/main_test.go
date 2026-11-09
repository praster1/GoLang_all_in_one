package main

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestApplyAll(t *testing.T) {
	identity := func(ctx context.Context, value int) (int, error) { return value, ctx.Err() }
	got, err := applyAll(context.Background(), []int{3, 1, 2}, identity)
	if err != nil || !reflect.DeepEqual(got, []int{3, 1, 2}) {
		t.Fatal(got, err)
	}
	got, err = applyAll(context.Background(), nil, identity)
	if err != nil || got == nil || len(got) != 0 {
		t.Fatal(got, err)
	}
	if _, err := applyAll(context.Background(), make([]int, 1000), identity); err != nil {
		t.Fatal(err)
	}
	for _, input := range []struct {
		values []int
		task   func(context.Context, int) (int, error)
	}{{nil, nil}, {make([]int, 1001), identity}} {
		got, err = applyAll(context.Background(), input.values, input.task)
		if err == nil || got != nil {
			t.Fatal(got, err)
		}
	}
}

func TestApplyAllStopsAndDiscards(t *testing.T) {
	for _, stopAfter := range []int{0, 1, 3} {
		ctx, cancel := context.WithCancel(context.Background())
		if stopAfter == 0 {
			cancel()
		}
		calls := 0
		got, err := applyAll(ctx, []int{1, 2, 3}, func(ctx context.Context, value int) (int, error) {
			calls++
			if calls == stopAfter {
				cancel()
			}
			return value, nil
		})
		cancel()
		if got != nil || !errors.Is(err, context.Canceled) || calls != stopAfter {
			t.Fatal(got, err, calls)
		}
	}
	want := errors.New("transform failed")
	calls := 0
	got, err := applyAll(context.Background(), []int{1, 2, 3}, func(context.Context, int) (int, error) {
		calls++
		if calls == 2 {
			return 0, want
		}
		return 10, nil
	})
	if got != nil || !errors.Is(err, want) || calls != 2 {
		t.Fatal(got, err, calls)
	}
}

func Example_main() {
	main()
	// Output:
	// 정상 결과: [2 4 6]
	// 실패 결과 없음: true
	// 취소 확인: true
}
