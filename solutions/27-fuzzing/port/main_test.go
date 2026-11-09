package main

import (
	"strconv"
	"testing"
)

func TestPort(t *testing.T) {
	for _, tc := range []struct {
		text string
		want uint16
		bad  bool
	}{
		{"1", 1, false}, {"65535", 65535, false}, {"08080", 8080, false},
		{"", 0, true}, {"0", 0, true}, {"65536", 0, true}, {"+80", 0, true}, {" 80", 0, true}, {"８０", 0, true},
	} {
		got, err := parsePort(tc.text)
		if (err != nil) != tc.bad || got != tc.want {
			t.Fatalf("%q: %d %v", tc.text, got, err)
		}
	}
}

func FuzzPort(f *testing.F) {
	for _, seed := range []string{"", "0", "1", "65535", "65536", "08080"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, text string) {
		if len(text) > 1024 {
			t.Skip()
		}
		got, err := parsePort(text)
		if err != nil {
			return
		}
		if got == 0 {
			t.Fatal("zero accepted")
		}
		canonical := strconv.FormatUint(uint64(got), 10)
		again, err := parsePort(canonical)
		if err != nil || again != got {
			t.Fatal("round trip")
		}
	})
}

func FuzzValidPort(f *testing.F) {
	f.Add(uint16(1))
	f.Add(uint16(65535))
	f.Fuzz(func(t *testing.T, value uint16) {
		if value == 0 {
			return
		}
		got, err := parsePort(strconv.FormatUint(uint64(value), 10))
		if err != nil || got != value {
			t.Fatal("valid value rejected")
		}
	})
}

func Example_main() {
	main()
	// Output:
	// 포트: 8080
}
