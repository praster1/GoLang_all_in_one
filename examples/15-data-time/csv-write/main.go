package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
)

type Session struct {
	Name    string
	Minutes int
}

func writeSessions(destination io.Writer, sessions []Session) error {
	writer := csv.NewWriter(destination)
	if err := writer.Write([]string{"name", "minutes"}); err != nil {
		return err
	}
	for _, session := range sessions {
		record := []string{session.Name, strconv.Itoa(session.Minutes)}
		if err := writer.Write(record); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}

func run() error {
	sessions := []Session{
		{Name: "Go, basics", Minutes: 25},
		{Name: `문자열 "연습"`, Minutes: 35},
	}
	var output bytes.Buffer
	if err := writeSessions(&output, sessions); err != nil {
		return fmt.Errorf("CSV 쓰기: %w", err)
	}
	fmt.Print(output.String())
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
