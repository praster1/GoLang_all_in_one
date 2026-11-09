package main

import "testing"

func TestGreeting(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "Go", input: "Go", want: "안녕하세요, Go!"},
		{name: "Korean name", input: "민수", want: "안녕하세요, 민수!"},
		{name: "surrounding space", input: "  민수  ", want: "안녕하세요, 민수!"},
		{name: "empty name", input: "", want: "안녕하세요, Go!"},
		{name: "whitespace only", input: " \t\n", want: "안녕하세요, Go!"},
		{name: "internal space", input: "Go learner", want: "안녕하세요, Go learner!"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := greeting(tt.input); got != tt.want {
				t.Errorf("greeting(%q) = %q; want %q", tt.input, got, tt.want)
			}
		})
	}
}

func Example_main() {
	main()
	// Output:
	// 안녕하세요, Go!
}
