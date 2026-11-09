package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCopyFile(t *testing.T) {
	directory := t.TempDir()
	source := filepath.Join(directory, "source")
	destination := filepath.Join(directory, "copy")
	content := []byte("Go\n한글\n")
	if err := os.WriteFile(source, content, 0o600); err != nil {
		t.Fatal(err)
	}
	n, err := copyFile(source, destination)
	if err != nil || n != int64(len(content)) {
		t.Fatalf("copy = %d, %v", n, err)
	}
	data, err := os.ReadFile(destination)
	if err != nil || string(data) != string(content) {
		t.Fatalf("content = %q, %v", data, err)
	}
	if err := os.Remove(destination); err != nil {
		t.Fatalf("destination still open: %v", err)
	}
}
func TestCopyFilePreservesExistingDestination(t *testing.T) {
	directory := t.TempDir()
	source := filepath.Join(directory, "source")
	destination := filepath.Join(directory, "copy")
	for path, text := range map[string]string{source: "new", destination: "keep"} {
		if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := copyFile(source, destination); !errors.Is(err, os.ErrExist) {
		t.Fatalf("existing destination error = %v", err)
	}
	data, err := os.ReadFile(destination)
	if err != nil || string(data) != "keep" {
		t.Fatalf("destination = %q, %v", data, err)
	}
	if _, err := copyFile(source, source); !errors.Is(err, os.ErrExist) {
		t.Fatalf("same path error = %v", err)
	}
}
func TestCopyFileMissingSourceCreatesNothing(t *testing.T) {
	directory := t.TempDir()
	destination := filepath.Join(directory, "copy")
	if _, err := copyFile(filepath.Join(directory, "missing"), destination); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("source error = %v", err)
	}
	if _, err := os.Stat(destination); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("destination unexpectedly exists: %v", err)
	}
}

func TestCopyFileEmpty(t *testing.T) {
	directory := t.TempDir()
	source := filepath.Join(directory, "empty")
	destination := filepath.Join(directory, "copy")
	if err := os.WriteFile(source, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if written, err := copyFile(source, destination); written != 0 || err != nil {
		t.Fatalf("empty copy = (%d, %v)", written, err)
	}
	data, err := os.ReadFile(destination)
	if err != nil || len(data) != 0 {
		t.Fatalf("empty destination = (%q, %v)", data, err)
	}
}

func Example_main() {
	main()
	// Output:
	// 복사한 바이트: 3
	// Go
}
