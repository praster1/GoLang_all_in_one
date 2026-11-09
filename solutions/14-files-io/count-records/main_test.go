package main

import (
	"errors"
	"strings"
	"testing"
)

type failedReader struct{ err error }

func (r failedReader) Read([]byte) (int, error) { return 0, r.err }

func TestCountRecords(t *testing.T) {
	cases := []struct {
		input string
		want  int
	}{
		{"", 0}, {"\n # 주석\r\n\t\n", 0},
		{"Go\r\n테스트", 2}, {"값 # 본문\n", 1},
	}
	for _, tc := range cases {
		got, err := countRecords(strings.NewReader(tc.input))
		if err != nil || got != tc.want {
			t.Errorf("countRecords(%q) = (%d, %v), want %d", tc.input, got, err, tc.want)
		}
	}
}
func TestCountRecordsReadError(t *testing.T) {
	failure := errors.New("reader failed")
	if count, err := countRecords(failedReader{failure}); count != 0 || !errors.Is(err, failure) {
		t.Fatalf("count/error = %d/%v", count, err)
	}
}

func TestCountRecordsOversizedToken(t *testing.T) {
	if count, err := countRecords(strings.NewReader(strings.Repeat("x", 128*1024))); count != 0 || err == nil {
		t.Fatalf("oversized token = (%d, %v)", count, err)
	}
}

func Example_main() {
	main()
	// Output:
	// 레코드: 2
}
