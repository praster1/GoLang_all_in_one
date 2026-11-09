package main

import (
	"errors"
	"strings"
	"testing"
)

func TestReadSessions(t *testing.T) {
	for _, input := range []string{
		"name,minutes\n\"Go, basics\",25\n", "name,minutes\r\n\"Go, basics\",25",
	} {
		got, err := readSessions(strings.NewReader(input))
		if err != nil || len(got) != 1 || got[0] != (Session{"Go, basics", 25}) {
			t.Fatalf("%q => %#v, %v", input, got, err)
		}
	}
	got, err := readSessions(strings.NewReader("name,minutes\n"))
	if err != nil || got == nil || len(got) != 0 {
		t.Fatalf("header only = %#v, %v", got, err)
	}
	for _, input := range []string{
		"", "title,minutes\n", "name,minutes\nGo\n", "name,minutes\nGo,1,2\n",
		"name,minutes\nGo,x\n", "name,minutes\nGo,0\n", "name,minutes\nGo,121\n",
		"name,minutes\n ,25\n", "name,minutes\n\"Go,25\n",
	} {
		got, err := readSessions(strings.NewReader(input))
		if err == nil || got != nil {
			t.Errorf("%q => %#v, %v", input, got, err)
		}
	}
	for _, count := range []int{1000, 1001} {
		got, err := readSessions(strings.NewReader("name,minutes\n" + strings.Repeat("Go,1\n", count)))
		if count == 1000 && (err != nil || len(got) != count) {
			t.Fatalf("1000 records = %d, %v", len(got), err)
		}
		if count == 1001 && (err == nil || got != nil) {
			t.Fatalf("1001 records = %d, %v", len(got), err)
		}
	}
}

type failedReader struct{ err error }

func (reader failedReader) Read([]byte) (int, error) { return 0, reader.err }

func TestCSVReadError(t *testing.T) {
	want := errors.New("read failure")
	_, err := readSessions(failedReader{want})
	if !errors.Is(err, want) {
		t.Fatalf("error = %v", err)
	}
}

func Example_main() {
	main()
	// Output:
	// Go, basics: 25분
	// 테스트: 35분
	// 합계: 60분
}
