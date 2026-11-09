package main

import (
	"reflect"
	"testing"
)

func TestFirstSquares(t *testing.T) {
	for _, tt := range []struct {
		count, take int
		want        []int
	}{{0, 0, []int{}}, {1000, 0, []int{}}, {1000, 3, []int{1, 4, 9}}, {3, 3, []int{1, 4, 9}}} {
		got, err := firstSquares(tt.count, tt.take)
		if err != nil || !reflect.DeepEqual(got, tt.want) {
			t.Fatal(got, err)
		}
	}
	for _, tt := range [][2]int{{-1, 0}, {1001, 1}, {2, 3}, {2, -1}} {
		if got, err := firstSquares(tt[0], tt[1]); got != nil || err == nil {
			t.Fatal(got, err)
		}
	}
}
