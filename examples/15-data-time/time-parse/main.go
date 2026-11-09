package main

import (
	"fmt"
	"os"
	"time"
)

func run() error {
	start, err := time.Parse(time.RFC3339, "2026-11-09T09:30:00+09:00")
	if err != nil {
		return err
	}
	fmt.Println("원래 표시:", start.Format(time.RFC3339))
	fmt.Println("UTC 표시:", start.UTC().Format(time.RFC3339))
	zone := time.FixedZone("UTC+9", 9*60*60)
	local, err := time.ParseInLocation("2006-01-02 15:04", "2026-11-09 09:30", zone)
	if err != nil {
		return err
	}
	fmt.Println("같은 순간:", start.Equal(local))
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
