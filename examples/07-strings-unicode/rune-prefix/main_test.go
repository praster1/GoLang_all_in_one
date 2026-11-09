package main

import "testing"

func TestPrefixRunes(t *testing.T) {
	for _, tc := range []struct {
		text  string
		count int
		want  string
		ok    bool
	}{
		{"", 0, "", true}, {"Go한글", 3, "Go한", true}, {"한글", 1, "한", true},
		{"한글", 9, "한글", true}, {"한글", 0, "", true}, {"한글", -1, "", false},
		{"\xff", 1, "", false}, {"\xff", 0, "", false}, {"e\u0301", 1, "e", true},
	} {
		got, ok := prefixRunes(tc.text, tc.count)
		if got != tc.want || ok != tc.ok {
			t.Fatalf("%q,%d got %q,%t", tc.text, tc.count, got, ok)
		}
	}
}

func Example_main() {
	main()
	// Output:
	// 앞 3룬: "Go한", 성공: true
	// 앞 0룬: "", 성공: true
}
