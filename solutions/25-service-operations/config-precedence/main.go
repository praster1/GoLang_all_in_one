package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strconv"
)

func limitConfig(args []string, lookup func(string) (string, bool)) (int, error) {
	limit := 20
	if value, ok := lookup("APP_LIMIT"); ok {
		n, err := strconv.Atoi(value)
		if err != nil || n < 1 || n > 100 {
			return 0, errors.New("APP_LIMIT은 1~100 정수여야 합니다")
		}
		limit = n
	}
	flags := flag.NewFlagSet("limit", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.IntVar(&limit, "limit", limit, "페이지 크기")
	if err := flags.Parse(args); err != nil {
		return 0, err
	}
	if flags.NArg() != 0 {
		return 0, errors.New("위치 인자는 받지 않습니다")
	}
	if limit < 1 || limit > 100 {
		return 0, errors.New("limit는 1~100이어야 합니다")
	}
	return limit, nil
}
func main() {
	absent := func(string) (string, bool) { return "", false }
	env := func(string) (string, bool) { return "30", true }
	for _, tc := range []struct {
		args   []string
		lookup func(string) (string, bool)
	}{{nil, absent}, {nil, env}, {[]string{"-limit", "40"}, env}} {
		n, err := limitConfig(tc.args, tc.lookup)
		if err != nil {
			fmt.Println("오류:", err)
			return
		}
		fmt.Println("크기:", n)
	}
	_, err := limitConfig([]string{"-limit", "0"}, env)
	fmt.Println("범위 오류:", err != nil)
}
