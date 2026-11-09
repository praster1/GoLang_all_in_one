package main

import "testing"

func TestMovePoint(t *testing.T) {
	for _, tc := range []struct {
		start  Point
		dx, dy int
		want   Point
		ok     bool
	}{
		{Point{10, 20}, -3, 5, Point{7, 25}, true}, {Point{}, -1000, 1000, Point{-1000, 1000}, true},
		{Point{1000, 0}, 0, 0, Point{1000, 0}, true}, {Point{1000, 0}, 1, 0, Point{1000, 0}, false},
		{Point{0, -1000}, 0, -1, Point{0, -1000}, false}, {Point{1, 2}, 2, 1000, Point{1, 2}, false},
		{Point{}, 1001, 0, Point{}, false}, {Point{}, 0, -1001, Point{}, false}, {Point{1001, 0}, -1, 0, Point{1001, 0}, false},
	} {
		value := tc.start
		ok := movePoint(&value, tc.dx, tc.dy)
		if ok != tc.ok || value != tc.want {
			t.Fatalf("case %+v: %+v %t", tc, value, ok)
		}
	}
	if movePoint(nil, 0, 0) {
		t.Fatal("nil accepted")
	}
}

func Example_main() {
	main()
	// Output:
	// 이동: true, 좌표: (7, 25)
	// 범위 초과: false, 좌표: (7, 25)
}
