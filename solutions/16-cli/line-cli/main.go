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
	flags := flag.NewFlagSet("count-lines", flag.ContinueOnError)
	flags.SetOutput(stderr)
	limit := flags.Int("limit", 10, "허용할 최대 항목 수: 0~100")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if flags.NArg() != 0 || *limit < 0 || *limit > 100 {
		fmt.Fprintln(stderr, "오류: limit은 0~100이어야 합니다")
		return 2
	}
	scanner := bufio.NewScanner(source)
	scanner.Buffer(make([]byte, 1024), 128*1024)
	count := 0
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == "" {
			continue
		}
		count++
		if count > *limit {
			fmt.Fprintln(stderr, "오류: 항목 수 제한을 넘었습니다")
			return 2
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(stderr, "입력 오류:", err)
		return 1
	}
	if _, err := fmt.Fprintf(stdout, "항목: %d\n", count); err != nil {
		fmt.Fprintln(stderr, "출력 오류:", err)
		return 1
	}
	return 0
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
