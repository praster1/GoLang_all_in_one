package main

import (
	"fmt"
	"os"
	"time"
)

func run() error {
	duration, err := time.ParseDuration("1h30m")
	if err != nil {
		return err
	}
	start := time.Date(2026, time.November, 9, 9, 30, 0, 0, time.UTC)
	end := start.Add(duration)
	fmt.Printf("기간: %.0f분\n", duration.Minutes())
	fmt.Println("종료:", end.Format(time.RFC3339))
	fmt.Println("차이:", end.Sub(start))
	fmt.Println("시작이 먼저:", start.Before(end))
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
