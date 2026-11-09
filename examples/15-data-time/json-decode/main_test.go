package main

import (
	"errors"
	"strings"
	"testing"
)

func TestDecodeTask(t *testing.T) {
	for _, input := range []string{
		`{"title":" Go ","minutes":1}`, `{"title":"Go","minutes":120}`,
	} {
		task, err := decodeTask(strings.NewReader(input))
		if err != nil || task.Title != "Go" {
			t.Fatalf("decodeTask(%q) = %#v, %v", input, task, err)
		}
	}
	for _, input := range []string{
		``, `{`, `null`, `{}`, `{"title":" ","minutes":25}`,
		`{"title":"Go","minutes":0}`, `{"title":"Go","minutes":121}`,
		`{"title":"Go","minutes":1.5}`, `{"title":"Go","minutes":"25"}`,
		`{"title":"Go","minutes":25,"extra":true}`,
		`{"title":"Go","minutes":25} {}`, `{"title":"Go","minutes":25} x`,
	} {
		task, err := decodeTask(strings.NewReader(input))
		if err == nil || task != (Task{}) {
			t.Errorf("invalid input %q = %#v, %v", input, task, err)
		}
	}
}

type failedReader struct{ err error }

func (reader failedReader) Read([]byte) (int, error) { return 0, reader.err }

func TestDecodeTaskReadError(t *testing.T) {
	want := errors.New("read failure")
	_, err := decodeTask(failedReader{want})
	if !errors.Is(err, want) {
		t.Fatalf("error = %v", err)
	}
}

func Example_main() {
	main()
	// Output:
	// Go: 25분
}
