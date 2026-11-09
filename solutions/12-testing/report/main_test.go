package main

import (
	"bytes"
	"errors"
	"slices"
	"strings"
	"testing"
)

func TestWriteRoster(t *testing.T) {
	names := []string{" 민수 ", "지수"}
	before := slices.Clone(names)
	var output bytes.Buffer
	if err := writeRoster(&output, names); err != nil {
		t.Fatal(err)
	}
	if got, want := output.String(), "학습자: 2\n- 민수\n- 지수\n"; got != want {
		t.Errorf("got=%q; want %q", got, want)
	}
	if !slices.Equal(names, before) {
		t.Error("input changed")
	}
}

func TestWriteRosterEmpty(t *testing.T) {
	var output bytes.Buffer
	if err := writeRoster(&output, nil); err != nil || output.String() != "학습자: 0\n" {
		t.Fatalf("output=%q err=%v", output.String(), err)
	}
}

func TestWriteRosterInvalid(t *testing.T) {
	for _, name := range []string{" ", "a\nb", "a\u0085b", strings.Repeat("가", 21), string([]byte{0xff})} {
		var output bytes.Buffer
		err := writeRoster(&output, []string{"정상", name})
		if !errors.Is(err, errName) || output.Len() != 0 {
			t.Fatalf("output=%q err=%v", output.String(), err)
		}
	}
}

func TestWriteRosterMaximumName(t *testing.T) {
	var output bytes.Buffer
	name := strings.Repeat("가", 20)
	if err := writeRoster(&output, []string{name}); err != nil || !strings.Contains(output.String(), name) {
		t.Fatalf("output=%q err=%v", output.String(), err)
	}
}

type failedWriter struct {
	err       error
	remaining int
}

func (w *failedWriter) Write(p []byte) (int, error) {
	if w.remaining == 0 {
		return 0, w.err
	}
	w.remaining--
	return len(p), nil
}

func TestWriteRosterWriteError(t *testing.T) {
	cause := errors.New("output closed")
	for _, remaining := range []int{0, 1} {
		if err := writeRoster(&failedWriter{err: cause, remaining: remaining}, []string{"지수"}); !errors.Is(err, cause) {
			t.Fatalf("remaining=%d error=%v", remaining, err)
		}
	}
}

func Example_main() {
	main()
	// Output:
	// 학습자: 2
	// - 민수
	// - 지수
}
