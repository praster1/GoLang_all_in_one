package main

import (
	"errors"
	"io"
	"strings"
	"testing"
)

type dataErrorReader struct{ err error }

func (r dataErrorReader) Read(p []byte) (int, error) {
	return copy(p, "Go"), r.err
}

func TestReadChunks(t *testing.T) {
	for _, input := range []string{"", "abcd", "abcde", "Go와 파일"} {
		got, err := readChunks(strings.NewReader(input))
		if err != nil || got != input {
			t.Errorf("readChunks(%q) = (%q, %v)", input, got, err)
		}
	}
}

func TestReadChunksProcessesDataBeforeError(t *testing.T) {
	text, err := readChunks(dataErrorReader{err: io.EOF})
	if text != "Go" || err != nil {
		t.Fatalf("data with EOF = (%q, %v)", text, err)
	}
	failure := errors.New("read failed")
	text, err = readChunks(dataErrorReader{err: failure})
	if text != "Go" || !errors.Is(err, failure) {
		t.Fatalf("partial data with error = (%q, %v)", text, err)
	}
}

func Example_main() {
	main()
	// Output:
	// Go와 파일
	// 바이트: 12
}
