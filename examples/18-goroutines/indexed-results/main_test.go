package main

import (
	"reflect"
	"testing"
)

func TestSquares(t *testing.T) {
	for _, c := range []struct{ values, want []int }{{nil, []int{}}, {[]int{3, 1, 2}, []int{9, 1, 4}}, {[]int{-1000, 0, 1000}, []int{1000000, 0, 1000000}}} {
		got, err := squares(c.values)
		if err != nil || !reflect.DeepEqual(got, c.want) {
			t.Fatal(got, err)
		}
	}
	if _, err := squares(make([]int, 16)); err != nil {
		t.Fatal(err)
	}
	for _, values := range [][]int{{-1001}, {1001}, make([]int, 17)} {
		got, err := squares(values)
		if err == nil || got != nil {
			t.Fatal(got, err)
		}
	}
}

func Example_main() {
	main()
	// Output:
	// 입력 순서의 결과: [9 1 4]
}
