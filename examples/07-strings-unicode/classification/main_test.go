package main

import "testing"

func TestClassify(t *testing.T) {
	for _, tc := range []struct {
		text string
		want [4]int
	}{
		{"", [4]int{}}, {"Go 한글 42!", [4]int{4, 2, 2, 1}},
		{"한１\t!", [4]int{1, 1, 1, 1}}, {"e\u0301", [4]int{1, 0, 0, 1}},
	} {
		a, b, c, d := classify(tc.text)
		if [4]int{a, b, c, d} != tc.want {
			t.Fatalf("%q: %d %d %d %d", tc.text, a, b, c, d)
		}
	}
}

func Example_main() {
	main()
	// Output:
	// 글자: 4, 숫자: 2, 공백: 2, 기타: 1
}
