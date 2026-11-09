package main

import "testing"

func TestSum(t *testing.T) {
	tests := []struct {
		name    string
		numbers []int
		want    int
	}{
		{name: "no_arguments", numbers: nil, want: 0},
		{name: "one_argument", numbers: []int{7}, want: 7},
		{name: "multiple_arguments", numbers: []int{10, 20, 30}, want: 60},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sum(tt.numbers...); got != tt.want {
				t.Fatalf("sum(%v) = %d, want %d", tt.numbers, got, tt.want)
			}
		})
	}
}

func Example_main() {
	main()
	// Output:
	// 인자 없음: 0
	// 세 값의 합: 60
}
