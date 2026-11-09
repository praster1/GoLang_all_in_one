package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("repeat", flag.ContinueOnError)
	flags.SetOutput(stderr)
	name := flags.String("name", "Go", "인사할 이름")
	times := flags.Int("times", 1, "반복 횟수: 1~5")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	value := strings.TrimSpace(*name)
	if flags.NArg() != 0 || value == "" || *times < 1 || *times > 5 {
		fmt.Fprintln(stderr, "오류: 이름과 횟수(1~5)를 확인하세요")
		return 2
	}
	for range *times {
		if _, err := fmt.Fprintf(stdout, "안녕하세요, %s!\n", value); err != nil {
			fmt.Fprintln(stderr, "출력 오류:", err)
			return 1
		}
	}
	return 0
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}
