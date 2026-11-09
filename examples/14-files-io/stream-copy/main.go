package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

func run() error {
	source := strings.NewReader("Go\nstream\n")
	var destination bytes.Buffer
	copied, err := io.Copy(&destination, source)
	if err != nil {
		return fmt.Errorf("복사: %w", err)
	}
	_, err = fmt.Printf("복사한 바이트: %d\n%s", copied, destination.String())
	return err
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
