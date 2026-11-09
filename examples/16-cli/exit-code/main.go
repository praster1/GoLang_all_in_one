package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
)

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "사용법: days 0~7")
		return 2
	}
	days, err := strconv.Atoi(args[0])
	if err != nil || days < 0 || days > 7 {
		fmt.Fprintln(stderr, "오류: 일수는 0~7 정수여야 합니다")
		return 2
	}
	if _, err := fmt.Fprintf(stdout, "학습일: %d\n", days); err != nil {
		fmt.Fprintln(stderr, "출력 오류:", err)
		return 1
	}
	return 0
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}
