package main

import "testing"

func TestDivide(t *testing.T) {
	tests := []struct {
		name          string
		dividend      int
		divisor       int
		wantQuotient  int
		wantRemainder int
		wantOK        bool
	}{
		{name: "positive_remainder", dividend: 17, divisor: 5, wantQuotient: 3, wantRemainder: 2, wantOK: true},
		{name: "zero_dividend", dividend: 0, divisor: 5, wantQuotient: 0, wantRemainder: 0, wantOK: true},
		{name: "divisor_larger", dividend: 3, divisor: 5, wantQuotient: 0, wantRemainder: 3, wantOK: true},
		{name: "exact_division", dividend: 10, divisor: 5, wantQuotient: 2, wantRemainder: 0, wantOK: true},
		{name: "negative_dividend", dividend: -17, divisor: 5, wantQuotient: 0, wantRemainder: 0, wantOK: false},
		{name: "zero_divisor", dividend: 17, divisor: 0, wantQuotient: 0, wantRemainder: 0, wantOK: false},
		{name: "negative_divisor", dividend: 17, divisor: -5, wantQuotient: 0, wantRemainder: 0, wantOK: false},
		{name: "zero_over_zero", dividend: 0, divisor: 0, wantQuotient: 0, wantRemainder: 0, wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quotient, remainder, ok := divide(tt.dividend, tt.divisor)
			if quotient != tt.wantQuotient || remainder != tt.wantRemainder || ok != tt.wantOK {
				t.Fatalf("divide(%d, %d) = (%d, %d, %t), want (%d, %d, %t)", tt.dividend, tt.divisor, quotient, remainder, ok, tt.wantQuotient, tt.wantRemainder, tt.wantOK)
			}
		})
	}
}

func Example_main() {
	main()
	// Output:
	// 몫: 3, 나머지: 2
	// 0으로 나누기 가능: false
}
