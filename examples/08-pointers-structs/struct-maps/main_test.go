package main

import "testing"

func TestAddStock(t *testing.T) {
	for _, tc := range []struct {
		start, amount, want int
		ok                  bool
	}{{3, 2, 5, true}, {99, 1, 100, true}, {100, 1, 100, false}, {0, 100, 100, true}, {1, 0, 1, false}, {1, -1, 1, false}, {1, 101, 1, false}, {-1, 1, -1, false}} {
		stock := map[string]Stock{"pen": {Quantity: tc.start}}
		if ok := addStock(stock, "pen", tc.amount); ok != tc.ok || stock["pen"].Quantity != tc.want {
			t.Fatalf("case %+v: %+v, %t", tc, stock, ok)
		}
	}
	stock := map[string]Stock{}
	if addStock(stock, "missing", 1) || len(stock) != 0 || addStock(nil, "pen", 1) {
		t.Fatal("missing/nil accepted")
	}
}

func Example_main() {
	main()
	// Output:
	// 입고: true
	// 맵 수량: 5
	// 복사본과 맵: 6 5
	// 없는 상품: false
}
