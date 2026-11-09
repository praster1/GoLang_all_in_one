package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/praster1/GoLang_all_in_one/examples/31-concurrent-batch/batch"
)

func readLines(input io.Reader) ([]string, error) {
	scanner := bufio.NewScanner(input)
	scanner.Buffer(make([]byte, 4096), batch.MaxItemBytes+2)
	lines := []string{}
	for scanner.Scan() {
		if len(lines) == batch.MaxItems {
			return nil, fmt.Errorf("입력은 %d줄까지 허용합니다", batch.MaxItems)
		}
		if len(scanner.Text()) > batch.MaxItemBytes {
			return nil, fmt.Errorf("한 줄은 %d바이트까지 허용합니다", batch.MaxItemBytes)
		}
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

func run(parent context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("batch", flag.ContinueOnError)
	flags.SetOutput(stderr)
	workers := flags.Int("workers", 3, "동시에 처리할 작업자 수 (1~32)")
	attempts := flags.Int("attempts", 3, "항목별 최대 시도 횟수 (1~5)")
	retryDelay := flags.Duration("retry-delay", 0, "재시도 기본 대기 시간 (0~1s)")
	timeout := flags.Duration("timeout", 3*time.Second, "배치 처리 제한 시간 (0 초과, 최대 1m)")
	inputPath := flags.String("input", "", "입력 파일; - 는 표준 입력, 생략하면 내장 예제")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	cfg := batch.Config{Workers: *workers, Attempts: *attempts, RetryDelay: *retryDelay}
	if flags.NArg() != 0 || *timeout <= 0 || *timeout > time.Minute {
		fmt.Fprintln(stderr, "위치 인자는 받지 않으며 timeout은 0 초과 1m 이하여야 합니다")
		return 2
	}
	if err := batch.ValidateConfig(cfg); err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	inputs := []string{"go", "retry:channels", "한글"}
	if *inputPath != "" {
		var err error
		if *inputPath == "-" {
			inputs, err = readLines(stdin)
		} else {
			var file *os.File
			file, err = os.Open(*inputPath)
			if err == nil {
				inputs, err = readLines(file)
				err = errors.Join(err, file.Close())
			}
		}
		if err != nil {
			fmt.Fprintln(stderr, "입력 오류:", err)
			return 1
		}
	}
	ctx, cancel := context.WithTimeout(parent, *timeout)
	defer cancel()
	// Each job index belongs to one worker; retries for that index are sequential.
	calls := make([]int, len(inputs))
	process := func(ctx context.Context, job batch.Job) (string, error) {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		calls[job.Index]++
		value := strings.TrimSpace(job.Input)
		if value == "" {
			return "", errors.New("empty item")
		}
		if strings.HasPrefix(value, "fail:") {
			return "", errors.New("permanent example failure")
		}
		if strings.HasPrefix(value, "retry:") {
			if calls[job.Index] == 1 {
				return "", fmt.Errorf("first attempt: %w", batch.ErrTemporary)
			}
			value = strings.TrimSpace(strings.TrimPrefix(value, "retry:"))
			if value == "" {
				return "", errors.New("empty retry item")
			}
		}
		return strings.ToUpper(value), nil
	}
	results, err := batch.Run(ctx, inputs, cfg, process)
	if results == nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if writeErr := batch.WriteJSON(stdout, results); writeErr != nil {
		fmt.Fprintln(stderr, "출력 오류:", writeErr)
		return 1
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return 130
		}
		return 1
	}
	return 0
}

// restoreOnCancel releases signal interception after the first interrupt.
// A second interrupt can then terminate even if a plain Reader is still blocked.
func restoreOnCancel(ctx context.Context, stop func()) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		<-ctx.Done()
		stop()
		close(done)
	}()
	return done
}

func realMain() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	restored := restoreOnCancel(ctx, stop)
	defer func() {
		stop()
		<-restored
	}()
	return run(ctx, os.Args[1:], os.Stdin, os.Stdout, os.Stderr)
}
func main() { os.Exit(realMain()) }
