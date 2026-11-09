package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func copyFile(sourcePath, destinationPath string) (written int64, err error) {
	source, err := os.Open(sourcePath)
	if err != nil {
		return 0, fmt.Errorf("원본 열기: %w", err)
	}
	defer func() {
		err = errors.Join(err, source.Close())
	}()

	destination, err := os.OpenFile(destinationPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return 0, fmt.Errorf("대상 만들기: %w", err)
	}
	defer func() {
		err = errors.Join(err, destination.Close())
	}()
	written, err = io.Copy(destination, source)
	return written, err
}

func run() (err error) {
	directory, err := os.MkdirTemp("", "go-copy-")
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, os.RemoveAll(directory))
	}()
	source := filepath.Join(directory, "source.txt")
	destination := filepath.Join(directory, "copy.txt")
	if err := os.WriteFile(source, []byte("Go\n"), 0o600); err != nil {
		return err
	}
	written, err := copyFile(source, destination)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(destination)
	if err != nil {
		return err
	}
	_, err = fmt.Printf("복사한 바이트: %d\n%s", written, data)
	return err
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
