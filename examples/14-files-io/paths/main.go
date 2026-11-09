package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func run() (err error) {
	directory, err := os.MkdirTemp("", "go-paths-")
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, os.RemoveAll(directory))
	}()

	for _, name := range []string{"zeta.txt", "alpha.txt"} {
		if err := os.WriteFile(filepath.Join(directory, name), nil, 0o600); err != nil {
			return err
		}
	}
	if err := os.Mkdir(filepath.Join(directory, "notes"), 0o700); err != nil {
		return err
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		fmt.Printf("%s 디렉터리=%t\n", entry.Name(), entry.IsDir())
	}
	fmt.Println("표시용 경로:", filepath.ToSlash(filepath.Join("data", "notes.txt")))
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
