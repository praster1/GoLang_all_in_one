package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

func resolveName(args []string, lookup func(string) (string, bool), diagnostics io.Writer) (string, error) {
	flags := flag.NewFlagSet("name", flag.ContinueOnError)
	flags.SetOutput(diagnostics)
	name := flags.String("name", "", "환경 변수보다 우선하는 이름")
	if err := flags.Parse(args); err != nil {
		return "", err
	}
	if flags.NArg() != 0 {
		return "", errors.New("위치 인자는 허용하지 않습니다")
	}
	specified := false
	flags.Visit(func(item *flag.Flag) {
		if item.Name == "name" {
			specified = true
		}
	})
	value := "Go"
	if specified {
		value = *name
	} else if environment, exists := lookup("STUDY_NAME"); exists {
		value = environment
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return "", errors.New("이름은 비어 있을 수 없습니다")
	}
	return value, nil
}

func run(args []string, lookup func(string) (string, bool), stdout, stderr io.Writer) int {
	name, err := resolveName(args, lookup, stderr)
	if errors.Is(err, flag.ErrHelp) {
		return 0
	}
	if err != nil {
		fmt.Fprintln(stderr, "설정 오류:", err)
		return 2
	}
	if _, err := fmt.Fprintln(stdout, "이름:", name); err != nil {
		fmt.Fprintln(stderr, "출력 오류:", err)
		return 1
	}
	return 0
}

func main() {
	os.Exit(run(os.Args[1:], os.LookupEnv, os.Stdout, os.Stderr))
}
