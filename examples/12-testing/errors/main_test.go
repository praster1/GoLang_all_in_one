package main

import (
	"errors"
	"testing"
)

func TestParseCount(t *testing.T) {
	for _, tc := range []struct {
		text string
		want int
	}{{"1", 1}, {"0012", 12}, {"100", 100}} {
		got, err := parseCount(tc.text)
		if err != nil || got != tc.want {
			t.Errorf("parseCount(%q) = %d, %v; want %d, nil", tc.text, got, err, tc.want)
		}
	}
}

func TestParseCountError(t *testing.T) {
	for _, text := range []string{"", "0", "101", "-1", "+1", " 1", "１", "999999999999999999999"} {
		t.Run(text, func(t *testing.T) {
			got, err := parseCount(text)
			if !errors.Is(err, errCount) {
				t.Fatalf("error = %v; want errCount", err)
			}
			var detail *countError
			if !errors.As(err, &detail) {
				t.Fatalf("error type = %T; want *countError", err)
			}
			if got != 0 || detail.Input != text {
				t.Errorf("got=%d input=%q; want 0, %q", got, detail.Input, text)
			}
		})
	}
}

func Example_main() {
	main()
	// Output:
	// 수량: 12
}
