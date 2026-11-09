package main

import "testing"

func TestCounterAdd(t *testing.T) {
	for _, tc := range []struct {
		value, amount, want int
		ok                  bool
	}{{0, 0, 0, true}, {0, 100, 100, true}, {100, 1, 100, false}, {10, -1, 10, false}, {-1, 1, -1, false}, {101, 0, 101, false}} {
		c := Counter{Value: tc.value}
		ok := c.Add(tc.amount)
		if ok != tc.ok || c.Value != tc.want {
			t.Fatalf("%+v: %+v %t", tc, c, ok)
		}
	}
	var c *Counter
	if c.Add(1) {
		t.Fatal("nil accepted")
	}
}

func Example_main() {
	main()
	// Output:
	// 값 변수 호출: true 3
	// 포인터 호출: true 5
	// nil 수신자: false
}
