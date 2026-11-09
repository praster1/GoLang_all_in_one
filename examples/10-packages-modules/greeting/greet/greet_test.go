package greet

import "testing"

func TestHello(t *testing.T) {
	for _, tc := range []struct{ name, want string }{{" Go ", "안녕하세요, Go!"}, {"\t\n", "안녕하세요, 방문자!"}, {"한글", "안녕하세요, 한글!"}} {
		if got := Hello(tc.name); got != tc.want {
			t.Fatalf("Hello(%q)=%q want=%q", tc.name, got, tc.want)
		}
	}
}
