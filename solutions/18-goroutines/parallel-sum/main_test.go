package main

import (
	"testing"
)

func TestSumParts(t *testing.T) {
	for workers := 1; workers <= 8; workers++ {
		for _, c := range []struct {
			values []int
			want   int
		}{{nil, 0}, {[]int{10}, 10}, {[]int{10, 20, 30, 40, 50}, 150}} {
			got, err := sumParts(c.values, workers)
			if err != nil || got != c.want {
				t.Fatal(workers, got, err)
			}
		}
	}
	maximum := make([]int, 1000)
	for i := range maximum {
		maximum[i] = 1000
	}
	if got, err := sumParts(maximum, 8); err != nil || got != 1000000 {
		t.Fatal(got, err)
	}
	for _, c := range []struct {
		values  []int
		workers int
	}{{nil, 0}, {nil, 9}, {[]int{-1}, 2}, {[]int{1001}, 2}, {make([]int, 1001), 2}} {
		if got, err := sumParts(c.values, c.workers); err == nil || got != 0 {
			t.Fatal(got, err)
		}
	}
}

func Example_main() {
	main()
	// Output:
	// 합계: 150
}
