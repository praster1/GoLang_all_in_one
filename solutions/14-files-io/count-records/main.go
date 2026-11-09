package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

func countRecords(source io.Reader) (int, error) {
	scanner := bufio.NewScanner(source)
	scanner.Buffer(make([]byte, 1024), 128*1024)
	count := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		count++
	}
	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("레코드 읽기: %w", err)
	}
	return count, nil
}

func main() {
	count, err := countRecords(strings.NewReader("# 메모\nGo\n\n테스트\n"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("레코드:", count)
}
