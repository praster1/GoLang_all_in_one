package main

import "testing"

func TestCounterContract(t *testing.T) {
	for _, tc := range []struct {
		start, amount, want int
		ok                  bool
	}{{0, 0, 0, true}, {0, 100, 100, true}, {100, 0, 100, true}, {60, 41, 60, false}, {3, -1, 3, false}, {-1, 1, -1, false}, {101, 0, 101, false}} {
		c := Counter{value: tc.start}
		ok := c.Add(tc.amount)
		if ok != tc.ok || c.Value() != tc.want {
			t.Fatalf("%+v: %d %t", tc, c.Value(), ok)
		}
	}
	c := Counter{value: 30}
	if !c.Reset() || c.Value() != 0 {
		t.Fatal("reset")
	}
	var missing *Counter
	if missing.Add(1) || missing.Reset() {
		t.Fatal("nil accepted")
	}
}

func Example_main() {
	main()
	// Output:
	// 추가: true 60
	// 한도 초과: false 60
	// 초기화: true 0
}
