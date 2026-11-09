package main

import "testing"

func TestWithdraw(t *testing.T) {
	for _, q := range []int{-1, 0, 6} {
		m := map[string]int{"pen": 5}
		if withdraw(m, "pen", q) || m["pen"] != 5 || len(m) != 1 {
			t.Fatal(q, m)
		}
	}
	m := map[string]int{"pen": 5, "book": 0}
	if !withdraw(m, "pen", 5) || m["pen"] != 0 {
		t.Fatal(m)
	}
	if _, ok := m["pen"]; !ok {
		t.Fatal("zero-stock entry must remain")
	}
	if withdraw(m, "book", 1) || withdraw(m, "missing", 1) || withdraw(nil, "pen", 1) {
		t.Fatal("invalid withdrawal")
	}
	if len(m) != 2 {
		t.Fatal("missing lookup inserted key")
	}
}

func Example_main() {
	main()
	// Output:
	// 첫 출고: true
	// 남은 pen: 2
	// 초과 출고: false
	// 없는 상품: false
	// 최종 pen: 2
}
