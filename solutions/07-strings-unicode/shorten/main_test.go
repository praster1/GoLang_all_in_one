package main

import "testing"

func TestShorten(t *testing.T) {
	for _, tc := range []struct {
		text  string
		limit int
		want  string
		ok    bool
	}{
		{"", 0, "", true}, {"한글", 2, "한글", true}, {"한글", 8, "한글", true},
		{"한글", 1, "한…", true}, {"한글", 0, "…", true}, {"Go", -1, "", false},
		{"\xff", 0, "", false}, {"e\u0301", 1, "e…", true}, {"🙂Go", 1, "🙂…", true},
	} {
		got, ok := shorten(tc.text, tc.limit)
		if got != tc.want || ok != tc.ok {
			t.Fatalf("%q,%d -> %q,%t", tc.text, tc.limit, got, ok)
		}
	}
}

func Example_main() {
	main()
	// Output:
	// 요약: "안녕하세요…", 성공: true
}
