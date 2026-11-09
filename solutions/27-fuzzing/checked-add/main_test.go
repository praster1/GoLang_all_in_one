package main

import (
	"errors"
	"math/big"
	"testing"
)

func TestAddChecked(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	minInt := -maxInt - 1
	for _, tc := range []struct {
		a, b, want int
		bad        bool
	}{
		{0, 0, 0, false}, {-1, 1, 0, false}, {maxInt, 0, maxInt, false},
		{maxInt, 1, 0, true}, {minInt, -1, 0, true}, {minInt, maxInt, -1, false},
	} {
		got, err := addChecked(tc.a, tc.b)
		if errors.Is(err, errOverflow) != tc.bad || got != tc.want {
			t.Fatalf("got=%d err=%v", got, err)
		}
	}
}

func FuzzAddChecked(f *testing.F) {
	f.Add(10, 20)
	f.Add(-1, 1)
	f.Add(int(^uint(0)>>1), 1)
	f.Fuzz(func(t *testing.T, a, b int) {
		maxInt := int(^uint(0) >> 1)
		minInt := -maxInt - 1
		exact := new(big.Int).Add(big.NewInt(int64(a)), big.NewInt(int64(b)))
		overflow := exact.Cmp(big.NewInt(int64(minInt))) < 0 || exact.Cmp(big.NewInt(int64(maxInt))) > 0
		got, err := addChecked(a, b)
		if errors.Is(err, errOverflow) != overflow {
			t.Fatal("range classification")
		}
		if !overflow && int64(got) != exact.Int64() {
			t.Fatal("wrong sum")
		}
	})
}

func Example_main() {
	main()
	// Output:
	// 범위 오류: true
	// 합계: 30
}
