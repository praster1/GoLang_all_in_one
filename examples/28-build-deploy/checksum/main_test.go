package main

import (
	"errors"
	"strings"
	"testing"
)

type brokenReader struct{ err error }

func (r brokenReader) Read(p []byte) (int, error) { return 0, r.err }

func TestDigest(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"abc", "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"},
		{"", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
	} {
		got, err := digest(strings.NewReader(tc.in))
		if err != nil || got != tc.want {
			t.Fatalf("got=%s err=%v", got, err)
		}
	}
	want := errors.New("read failed")
	if _, err := digest(brokenReader{want}); !errors.Is(err, want) {
		t.Fatal("lost error")
	}
}

func Example_main() {
	main()
	// Output:
	// ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad
}
