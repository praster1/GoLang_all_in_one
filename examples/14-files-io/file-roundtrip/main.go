package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func run() (err error) {
	directory, err := os.MkdirTemp("", "go-file-")
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, os.RemoveAll(directory))
	}()

	path := filepath.Join(directory, "notes.txt")
	if err := os.WriteFile(path, []byte("Go\n파일\n"), 0o600); err != nil {
		return fmt.Errorf("쓰기: %w", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("읽기: %w", err)
	}
	_, err = fmt.Printf("읽은 바이트: %d\n%s", len(data), data)
	return err
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
