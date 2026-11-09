package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

func summarize(source io.Reader, destination io.Writer) error {
	scanner := bufio.NewScanner(source)
	scanner.Buffer(make([]byte, 1024), 128*1024)
	count := 0
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) != "" {
			count++
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("입력 읽기: %w", err)
	}
	_, err := fmt.Fprintf(destination, "학습 항목: %d\n", count)
	return err
}

func main() {
	if err := summarize(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
