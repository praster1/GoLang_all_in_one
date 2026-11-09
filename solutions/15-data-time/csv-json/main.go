package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type Session struct {
	Name    string `json:"name"`
	Minutes int    `json:"minutes"`
}

func convertCSV(source io.Reader, destination io.Writer) error {
	reader := csv.NewReader(source)
	reader.FieldsPerRecord = 2
	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("헤더 읽기: %w", err)
	}
	if header[0] != "name" || header[1] != "minutes" {
		return errors.New("헤더는 name,minutes여야 합니다")
	}
	sessions := make([]Session, 0)
	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("레코드 읽기: %w", err)
		}
		if len(sessions) == 1000 {
			return errors.New("레코드는 1000개까지 허용합니다")
		}
		minutes, err := strconv.Atoi(record[1])
		if err != nil {
			return fmt.Errorf("분 변환: %w", err)
		}
		name := strings.TrimSpace(record[0])
		if name == "" || minutes < 0 || minutes > 120 {
			return errors.New("이름과 0~120분이 필요합니다")
		}
		sessions = append(sessions, Session{Name: name, Minutes: minutes})
	}
	encoder := json.NewEncoder(destination)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(sessions); err != nil {
		return fmt.Errorf("JSON 쓰기: %w", err)
	}
	return nil
}

func run() error {
	var output bytes.Buffer
	input := "name,minutes\nGo,25\n테스트,35\n"
	if err := convertCSV(strings.NewReader(input), &output); err != nil {
		return err
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
