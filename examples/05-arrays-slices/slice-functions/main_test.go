package main

import (
	"slices"
	"testing"
)

func TestIncrementAndAppendContract(t *testing.T) {
	increment(nil)
	values := []int{-1, 0, 1}
	increment(values)
	if !slices.Equal(values, []int{0, 1, 2}) {
		t.Fatal(values)
	}
	out := addValue(values, 9)
	if len(values) != 3 || !slices.Equal(out, []int{0, 1, 2, 9}) {
		t.Fatalf("%v %v", values, out)
	}
	if got := addValue(nil, 4); !slices.Equal(got, []int{4}) {
		t.Fatal(got)
	}
}

func Example_main() {
	main()
	// Output:
	// 원소 변경: [11 21]
	// 결과를 버린 길이: 2
	// 반환값 대입: [11 21 40]
}
