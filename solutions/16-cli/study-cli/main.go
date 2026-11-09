package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

type Summary struct {
	Count        int `json:"count"`
	TotalMinutes int `json:"total_minutes"`
}

func run(args []string, source io.Reader, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("study", flag.ContinueOnError)
	flags.SetOutput(stderr)
	minutes := flags.Int("minutes", 25, "항목별 분: 1~120")
	jsonOutput := flags.Bool("json", false, "JSON으로 출력")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if flags.NArg() != 0 || *minutes < 1 || *minutes > 120 {
		fmt.Fprintln(stderr, "오류: minutes는 1~120이어야 합니다")
		return 2
	}
	data, err := io.ReadAll(io.LimitReader(source, 4097))
	if err != nil {
		fmt.Fprintln(stderr, "입력 오류:", err)
		return 1
	}
	if len(data) > 4096 {
		fmt.Fprintln(stderr, "오류: 입력은 4096바이트까지 허용합니다")
		return 2
	}
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	count := 0
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) != "" {
			count++
		}
		if count > 100 {
			fmt.Fprintln(stderr, "오류: 항목은 100개까지 허용합니다")
			return 2
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(stderr, "입력 오류:", err)
		return 1
	}
	summary := Summary{Count: count, TotalMinutes: count * *minutes}
	if *jsonOutput {
		err = json.NewEncoder(stdout).Encode(summary)
	} else {
		_, err = fmt.Fprintf(stdout, "항목: %d\n합계: %d분\n", summary.Count, summary.TotalMinutes)
	}
	if err != nil {
		fmt.Fprintln(stderr, "출력 오류:", err)
		return 1
	}
	return 0
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
