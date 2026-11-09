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
	flags := flag.NewFlagSet("greet", flag.ContinueOnError)
	flags.SetOutput(stderr)
	name := flags.String("name", "Go", "인사할 이름")
	upper := flags.Bool("upper", false, "이름의 영문자를 대문자로")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if flags.NArg() != 0 || strings.TrimSpace(*name) == "" {
		fmt.Fprintln(stderr, "오류: 빈 이름과 위치 인자는 허용하지 않습니다")
		return 2
	}
	value := strings.TrimSpace(*name)
	if *upper {
		value = strings.ToUpper(value)
	}
	if _, err := fmt.Fprintf(stdout, "안녕하세요, %s!\n", value); err != nil {
		fmt.Fprintln(stderr, "출력 오류:", err)
		return 1
	}
	return 0
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}
