package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestReadText(t *testing.T) {
	path := filepath.Join(t.TempDir(), "input.txt")
	if err := os.WriteFile(path, []byte("Go\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	data, err := readText(path)
	if err != nil || string(data) != "Go\n" {
		t.Fatalf("readText = (%q, %v)", data, err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatalf("remove after readText: %v", err)
	}
}

func TestReadTextMissingFile(t *testing.T) {
	_, err := readText(filepath.Join(t.TempDir(), "missing"))
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("error = %v, want os.ErrNotExist", err)
	}
}

func Example_main() {
	main()
	// Output:
	// 열고 읽고 닫기
}
