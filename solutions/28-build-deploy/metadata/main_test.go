package main

import (
	"strings"
	"testing"
)

func TestArtifactName(t *testing.T) {
	for _, tc := range []struct {
		r    release
		want string
		bad  bool
	}{
		{release{"hello", "1.0.0", "linux", "amd64"}, "hello-1.0.0-linux-amd64", false},
		{release{"hello", "dev-1", "windows", "arm64"}, "hello-dev-1-windows-arm64.exe", false},
		{release{"../bad", "1", "linux", "amd64"}, "", true},
		{release{"hello", "two words", "linux", "amd64"}, "", true},
		{release{"hello", "1", "other", "amd64"}, "", true},
		{release{"hello", "1", "linux", "386"}, "", true},
		{release{strings.Repeat("a", 64), "1", "linux", "amd64"}, strings.Repeat("a", 64) + "-1-linux-amd64", false},
		{release{strings.Repeat("a", 65), "1", "linux", "amd64"}, "", true},
	} {
		got, err := artifactName(tc.r)
		if (err != nil) != tc.bad || got != tc.want {
			t.Fatalf("got=%q err=%v", got, err)
		}
	}
}

func Example_main() {
	main()
	// Output:
	// hello-1.0.0-windows-amd64.exe
}
