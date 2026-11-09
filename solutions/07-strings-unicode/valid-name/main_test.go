package main

import "testing"

func TestValidName(t *testing.T) {
	for _, tc := range []struct {
		text string
		want bool
	}{
		{"", false}, {"a", false}, {"a1", true}, {"abcdefgh", true}, {"abcdefghi", false},
		{"한글_2", true}, {"2Go", false}, {"Go 학습", false}, {"a-b", false},
		{"a１", true}, {"e\u0301", false}, {"é1", true}, {"Go🙂", false}, {"a\xff", false},
	} {
		if got := validName(tc.text); got != tc.want {
			t.Fatalf("%q=%t", tc.text, got)
		}
	}
}

func Example_main() {
	main()
	// Output:
	// 한글_2: true
	// 2Go: false
	// Go 학습: false
}
