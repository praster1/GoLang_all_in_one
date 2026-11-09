package main

import "testing"

func TestShippingFee(t *testing.T) {
	tests := []struct {
		name        string
		grams, want int
		wantErr     bool
	}{
		{"minimum", 1, 3000, false}, {"first_edge", 1000, 3000, false},
		{"after_first", 1001, 5000, false}, {"second_edge", 5000, 5000, false},
		{"after_second", 5001, 8000, false}, {"maximum", 10000, 8000, false},
		{"zero", 0, 0, true}, {"negative", -1, 0, true}, {"too_heavy", 10001, 0, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := shippingFee(tc.grams)
			if (err != nil) != tc.wantErr {
				t.Fatalf("shippingFee(%d) error = %v; wantErr %v", tc.grams, err, tc.wantErr)
			}
			if got != tc.want {
				t.Errorf("shippingFee(%d) = %d; want %d", tc.grams, got, tc.want)
			}
		})
	}
}

func Example_main() {
	main()
	// Output:
	// 배송비: 5000
}
