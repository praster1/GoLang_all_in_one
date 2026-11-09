package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type Session struct {
	Name    string
	Minutes int
}

func readSessions(source io.Reader) ([]Session, error) {
	reader := csv.NewReader(source)
	reader.FieldsPerRecord = 2
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("CSV 헤더: %w", err)
	}
	if header[0] != "name" || header[1] != "minutes" {
		return nil, errors.New("헤더는 name,minutes여야 합니다")
	}
	sessions := make([]Session, 0)
	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			return sessions, nil
		}
		if err != nil {
			return nil, fmt.Errorf("CSV 레코드: %w", err)
		}
		if len(sessions) == 1000 {
			return nil, errors.New("레코드는 1000개까지 허용합니다")
		}
		name := strings.TrimSpace(record[0])
		minutes, err := strconv.Atoi(record[1])
		if err != nil {
			return nil, fmt.Errorf("분을 정수로 변환: %w", err)
		}
		if name == "" || minutes < 1 || minutes > 120 {
			return nil, errors.New("이름과 1~120분이 필요합니다")
		}
		sessions = append(sessions, Session{Name: name, Minutes: minutes})
	}
}

func run() error {
	input := "name,minutes\n\"Go, basics\",25\n테스트,35\n"
	sessions, err := readSessions(strings.NewReader(input))
	if err != nil {
		return err
	}
	total := 0
	for _, session := range sessions {
		fmt.Printf("%s: %d분\n", session.Name, session.Minutes)
		total += session.Minutes
	}
	fmt.Printf("합계: %d분\n", total)
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
