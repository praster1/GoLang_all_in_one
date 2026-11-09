package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

func writeNumbered(destination io.Writer, source io.Reader) error {
	scanner := bufio.NewScanner(source)
	scanner.Buffer(make([]byte, 1024), 128*1024)
	number := 0
	for scanner.Scan() {
		number++
		if _, err := fmt.Fprintf(destination, "%d: %s\n", number, scanner.Text()); err != nil {
			return fmt.Errorf("줄 %d 쓰기: %w", number, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("줄 읽기: %w", err)
	}
	return nil
}

func main() {
	if err := writeNumbered(os.Stdout, strings.NewReader("Go\n테스트\n끝")); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
