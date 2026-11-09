package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func openFixture(t *testing.T, text string) *os.File {
	t.Helper()
	path := filepath.Join(t.TempDir(), "label.txt")
	if err := os.WriteFile(path, []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := file.Close(); err != nil {
			t.Errorf("close fixture: %v", err)
		}
	})
	return file
}

func TestReadLabel(t *testing.T) {
	file := openFixture(t, "  한글\n")
	got, err := readLabel(file)
	if err != nil || got != "한글" {
		t.Fatalf("got=%q err=%v", got, err)
	}
}

func TestReadLabelBoundary(t *testing.T) {
	for _, tc := range []struct {
		name, text string
		bad        bool
	}{
		{"twenty", strings.Repeat("가", 20), false}, {"twenty_one", strings.Repeat("가", 21), true},
		{"empty", " \n", true}, {"multiple_lines", "a\nb", true}, {"invalid_utf8", string([]byte{0xff}), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := readLabel(openFixture(t, tc.text))
			if (err != nil) != tc.bad {
				t.Fatalf("got=%q err=%v", got, err)
			}
			if tc.bad && got != "" {
				t.Error("result on failure must be empty")
			}
		})
	}
}

type failedReader struct{ err error }

func (r failedReader) Read(p []byte) (int, error) { return 0, r.err }

func TestReadLabelReadError(t *testing.T) {
	cause := errors.New("disk read failed")
	got, err := readLabel(failedReader{cause})
	if got != "" || !errors.Is(err, cause) {
		t.Fatalf("got=%q err=%v", got, err)
	}
}

func Example_main() {
	main()
	// Output:
	// 라벨: 한글
}
