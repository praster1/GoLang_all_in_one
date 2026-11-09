package main

import "testing"

func TestSumEven(t *testing.T) {
	tests := []struct {
		name   string
		values []int
		want   int
	}{
		{"no arguments", nil, 0},
		{"one even number", []int{8}, 8},
		{"only odd numbers", []int{1, 3, 5}, 0},
		{"default input", []int{1, 2, 3, 4, 6}, 12},
		{"negative even number", []int{-2, 3, 4}, 2},
		{"negative odd and even numbers", []int{-4, -3, 2, 0}, -2},
		{"zero", []int{0}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sumEven(tt.values...); got != tt.want {
				t.Errorf("sumEven(%v) = %d, want %d", tt.values, got, tt.want)
			}
		})
	}
}

func Example_main() {
	main()
	// Output:
	// 짝수 합계: 12
}
