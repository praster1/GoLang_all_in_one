package main

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestTake(t *testing.T) {
	for _, limit := range []int{0, 1, 3} {
		in := make(chan int, 2)
		in <- 1
		in <- 2
		close(in)
		want := []int{1, 2}
		if limit < 2 {
			want = want[:limit]
		}
		got, err := take(context.Background(), in, limit)
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Fatal(got, err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if got, err := take(ctx, nil, 1); got != nil || !errors.Is(err, context.Canceled) {
		t.Fatal(got, err)
	}
	for _, limit := range []int{-1, 1001} {
		if got, err := take(context.Background(), nil, limit); got != nil || err == nil {
			t.Fatal(got, err)
		}
	}
}
