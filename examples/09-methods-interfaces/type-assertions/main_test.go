package main

import "testing"

func TestDescribeValue(t *testing.T) {
	for _, tc := range []struct {
		value any
		want  string
	}{{nil, "값 없음"}, {"", "문자열 \"\""}, {-2, "정수 -2"}, {int64(2), "다른 타입 int64"}, {false, "다른 타입 bool"}} {
		if got := describeValue(tc.value); got != tc.want {
			t.Fatalf("%T: %q", tc.value, got)
		}
	}
}

func Example_main() {
	main()
	// Output:
	// int 확인: 42 true
	// string 확인: "" false
	// 문자열 "Go"
	// 정수 7
	// 다른 타입 bool
	// 값 없음
}
