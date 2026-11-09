package main

import (
	"errors"
	"strings"
	"testing"
)

type failedReader struct{ err error }

func (r failedReader) Read([]byte) (int, error) { return 0, r.err }

func TestReadBounded(t *testing.T) {
	cases := []struct {
		text     string
		limit    int64
		tooLarge bool
	}{
		{"", 0, false}, {"x", 0, true},
		{"abcd", 5, false}, {"hello", 5, false}, {"hello!", 5, true},
		{"한", 3, false}, {"한", 2, true},
	}
	for _, tc := range cases {
		data, err := readBounded(strings.NewReader(tc.text), tc.limit)
		if tc.tooLarge {
			if data != nil || !errors.Is(err, errTooLarge) {
				t.Errorf("input %q: %q, %v", tc.text, data, err)
			}
		} else if err != nil || string(data) != tc.text {
			t.Errorf("input %q: %q, %v", tc.text, data, err)
		}
	}
}
func TestReadBoundedInvalidLimitAndReadFailure(t *testing.T) {
	for _, maximum := range []int64{-1, 1024*1024 + 1} {
		if _, err := readBounded(strings.NewReader(""), maximum); err == nil {
			t.Errorf("accepted %d", maximum)
		}
	}
	failure := errors.New("source failed")
	if _, err := readBounded(failedReader{failure}, 10); !errors.Is(err, failure) {
		t.Fatalf("error = %v", err)
	}
}

func Example_main() {
	main()
	// Output:
	// 내용: hello
	// 초과 확인: true
}
