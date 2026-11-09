package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

func run(args []string, source io.Reader, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("study-lines", flag.ContinueOnError)
	flags.SetOutput(stderr)
	prefix := flags.String("prefix", "학습", "각 항목 앞에 붙일 이름")
	limit := flags.Int("limit", 3, "허용할 최대 항목 수: 1~100")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if flags.NArg() != 0 || *limit < 1 || *limit > 100 || strings.TrimSpace(*prefix) == "" {
		fmt.Fprintln(stderr, "오류: prefix와 limit(1~100)을 확인하세요")
		return 2
	}
	scanner := bufio.NewScanner(source)
	scanner.Buffer(make([]byte, 1024), 128*1024)
	lines := make([]string, 0)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if len(lines) == *limit {
			fmt.Fprintln(stderr, "오류: 항목 수 제한을 넘었습니다")
			return 2
		}
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(stderr, "입력 오류:", err)
		return 1
	}
	for index, line := range lines {
		if _, err := fmt.Fprintf(stdout, "%s %d: %s\n", strings.TrimSpace(*prefix), index+1, line); err != nil {
			fmt.Fprintln(stderr, "출력 오류:", err)
			return 1
		}
	}
	return 0
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
