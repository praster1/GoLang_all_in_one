package main

import "testing"

func TestRectangle(t *testing.T) {
	for _, tc := range []struct {
		rectangle       Rectangle
		area, perimeter int
	}{{Rectangle{3, 4}, 12, 14}, {Rectangle{}, 0, 0}, {Rectangle{0, 5}, 0, 10}, {Rectangle{1000, 1000}, 1000000, 4000}} {
		before := tc.rectangle
		if got := tc.rectangle.Area(); got != tc.area {
			t.Fatalf("area %+v: %d", tc, got)
		}
		if got := tc.rectangle.Perimeter(); got != tc.perimeter {
			t.Fatalf("perimeter %+v: %d", tc, got)
		}
		if tc.rectangle != before {
			t.Fatal("receiver changed")
		}
	}
}

func Example_main() {
	main()
	// Output:
	// 넓이: 12
	// 둘레: 14
	// 원본: 3 4
}
