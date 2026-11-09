package main

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestCollect(t *testing.T) {
	got, err := collect(context.Background(), [][]int{{3, 1}, {2, 3}, nil})
	if err != nil || !reflect.DeepEqual(got, []int{1, 2, 3, 3}) {
		t.Fatal(got, err)
	}
	got, err = collect(context.Background(), nil)
	if err != nil || got == nil || len(got) != 0 {
		t.Fatal(got, err)
	}
	for _, parts := range [][][]int{make([][]int, 9), {make([]int, 1001)}} {
		if got, err := collect(context.Background(), parts); got != nil || err == nil {
			t.Fatal(got, err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if got, err := collect(ctx, nil); got != nil || !errors.Is(err, context.Canceled) {
		t.Fatal(got, err)
	}
}
