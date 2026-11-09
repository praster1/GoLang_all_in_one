package main

import (
	"reflect"
	"testing"
)

func TestOrdered(t *testing.T) {
	for _, order := range [][]int{{2, 0, 1}, {0, 1, 2}, {1, 2, 0}} {
		got, actual, err := ordered(order)
		if err != nil || !reflect.DeepEqual(actual, order) || !reflect.DeepEqual(got, []int{0, 10, 20}) {
			t.Fatal(got, actual, err)
		}
	}
	for _, order := range [][]int{{0, 0}, {-1}, {1}, make([]int, 17)} {
		if got, actual, err := ordered(order); got != nil || actual != nil || err == nil {
			t.Fatal(got, actual, err)
		}
	}
	if got, actual, err := ordered(nil); err != nil || len(got) != 0 || len(actual) != 0 {
		t.Fatal(got, actual, err)
	}
}
