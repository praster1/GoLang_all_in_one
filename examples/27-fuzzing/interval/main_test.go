package main

import (
	"fmt"
	"testing"
)

func TestInterval(t *testing.T) {
	for _, input := range []string{"", ":", "-1:2", "3:2", "0:10001", "1:2:3", " 1:2", "+1:2"} {
		if _, _, err := parseInterval(input); err == nil {
			t.Fatalf("accepted %q", input)
		}
	}
	for _, tc := range []struct {
		input     string
		low, high int
	}{{"0:0", 0, 0}, {"0002:10", 2, 10}, {"0:10000", 0, 10000}} {
		low, high, err := parseInterval(tc.input)
		if err != nil || low != tc.low || high != tc.high {
			t.Fatal("bad result")
		}
	}
}

func FuzzInterval(f *testing.F) {
	for _, seed := range []string{"", "0:0", "2:10", "0:10000", "1:2:3"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, input string) {
		if len(input) > 1024 {
			t.Skip()
		}
		low, high, err := parseInterval(input)
		if err != nil {
			return
		}
		if low < 0 || low > high || high > 10000 {
			t.Fatal("range contract")
		}
		again := fmt.Sprintf("%d:%d", low, high)
		l, h, e := parseInterval(again)
		if e != nil || l != low || h != high {
			t.Fatal("semantic round trip")
		}
	})
}

func Example_main() {
	main()
	// Output:
	// 범위: 2~10
}
