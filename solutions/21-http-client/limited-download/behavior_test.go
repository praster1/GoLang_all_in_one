package main

import (
	"errors"
	"io"
	"strings"
	"testing"
)

type failingReader struct{}

func (failingReader) Read(p []byte) (int, error) { return 0, io.ErrUnexpectedEOF }
func TestReadAtMost(t *testing.T) {
	for _, tc := range []struct {
		body  string
		limit int64
		want  string
		large bool
	}{
		{"", 0, "", false}, {"x", 0, "", true}, {"abcd", 4, "abcd", false}, {"abcde", 4, "", true}, {"가", 3, "가", false}, {"가", 2, "", true},
	} {
		got, err := readAtMost(strings.NewReader(tc.body), tc.limit)
		if tc.large {
			if !errors.Is(err, ErrTooLarge) {
				t.Errorf("body=%q err=%v", tc.body, err)
			}
		} else if err != nil || string(got) != tc.want {
			t.Errorf("body=%q got=%q err=%v", tc.body, got, err)
		}
	}
	for _, limit := range []int64{-1, 1025} {
		if _, err := readAtMost(strings.NewReader(""), limit); err == nil {
			t.Errorf("accepted limit=%d", limit)
		}
	}
	if _, err := readAtMost(failingReader{}, 4); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("reader error=%v", err)
	}
}
