package main

import (
	"testing"
	"time"
)

func TestWithinWindow(t *testing.T) {
	start := time.Date(2026, time.November, 9, 9, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	for _, test := range []struct {
		value time.Time
		want  bool
	}{
		{start.Add(-time.Nanosecond), false}, {start, true},
		{end.Add(-time.Nanosecond), true}, {end, false},
		{end.Add(time.Hour), false},
		{start.In(time.FixedZone("UTC+9", 9*60*60)), true},
	} {
		if got := withinWindow(test.value, start, end); got != test.want {
			t.Errorf("%s => %t; want %t", test.value, got, test.want)
		}
	}
	if withinWindow(start, start, start) || withinWindow(start, end, start) {
		t.Fatal("empty and reversed windows must reject")
	}
}

func Example_main() {
	main()
	// Output:
	// 2026-11-09T08:59:59.999999999Z false
	// 2026-11-09T09:00:00Z true
	// 2026-11-09T09:59:59.999999999Z true
	// 2026-11-09T10:00:00Z false
}
