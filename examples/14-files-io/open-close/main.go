package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func readText(path string) (data []byte, err error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("열기: %w", err)
	}
	defer func() {
		err = errors.Join(err, file.Close())
	}()
	data, err = io.ReadAll(file)
	return data, err
}

func run() (err error) {
	directory, err := os.MkdirTemp("", "go-open-")
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, os.RemoveAll(directory))
	}()

	path := filepath.Join(directory, "lesson.txt")
	if err := os.WriteFile(path, []byte("열고 읽고 닫기\n"), 0o600); err != nil {
		return err
	}
	data, err := readText(path)
	if err != nil {
		return err
	}
	_, err = fmt.Print(string(data))
	return err
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
