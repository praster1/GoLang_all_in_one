package main

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestSum(t *testing.T) {
	for _, test := range []struct {
		values []int
		want   int
	}{
		{nil, 0}, {[]int{0}, 0}, {[]int{10, 20, 30}, 60}, {[]int{1000}, 1000},
	} {
		got, err := sum(context.Background(), test.values)
		if err != nil || got != test.want {
			t.Fatal(got, err)
		}
	}
	maximum := make([]int, 1000)
	for i := range maximum {
		maximum[i] = 1000
	}
	if got, err := sum(context.Background(), maximum); err != nil || got != 1000000 {
		t.Fatal(got, err)
	}
	for _, values := range [][]int{{-1}, {1001}, {10, -1}, make([]int, 1001)} {
		if got, err := sum(context.Background(), values); err == nil || got != 0 {
			t.Fatal(got, err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for _, values := range [][]int{nil, {10}, {-1}} {
		if got, err := sum(ctx, values); got != 0 || !errors.Is(err, context.Canceled) {
			t.Fatal(got, err)
		}
	}
	expired, stop := context.WithDeadline(context.Background(), time.Unix(0, 0))
	defer stop()
	if _, err := sum(expired, []int{1}); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
}

func Example_main() {
	main()
	// Output:
	// 합계: 60
	// 취소 확인: true
}
