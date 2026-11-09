package main

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestPositive(t *testing.T) {
	for _, tt := range []struct{ input, want []int }{{nil, []int{}}, {[]int{-2, 0, 3, 1}, []int{3, 1}}, {[]int{-1, 0}, []int{}}} {
		got, err := positive(context.Background(), tt.input)
		if err != nil || !reflect.DeepEqual(got, tt.want) {
			t.Fatal(got, err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if got, err := positive(ctx, []int{1}); got != nil || !errors.Is(err, context.Canceled) {
		t.Fatal(got, err)
	}
	if got, err := positive(context.Background(), make([]int, 1001)); got != nil || err == nil {
		t.Fatal(got, err)
	}
}
