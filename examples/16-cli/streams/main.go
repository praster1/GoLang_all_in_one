package main

import (
	"fmt"
	"io"
	"os"
)

func writeResult(stdout, stderr io.Writer) error {
	if _, err := fmt.Fprintln(stderr, "처리: 2개"); err != nil {
		return err
	}
	_, err := fmt.Fprint(stdout, "Go\nTest\n")
	return err
}

func main() {
	if err := writeResult(os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "출력 오류:", err)
		os.Exit(1)
	}
}
