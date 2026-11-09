package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

func countNonEmpty(source io.Reader) (int, error) {
	scanner := bufio.NewScanner(source)
	scanner.Buffer(make([]byte, 1024), 128*1024)
	count := 0
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) != "" {
			count++
		}
	}
	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("줄 읽기: %w", err)
	}
	return count, nil
}

func main() {
	count, err := countNonEmpty(strings.NewReader("Go\n\n  \n한글"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("비어 있지 않은 줄:", count)
}
