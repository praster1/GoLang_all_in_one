package main

import "testing"

func TestVersionLine(t *testing.T) {
	if got := versionLine("1.0.0", "abc123"); got != "version=1.0.0 revision=abc123" {
		t.Fatal(got)
	}
}

func Example_main() {
	main()
	// Output:
	// version=dev revision=unknown
}
