package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

func courseName(lookup func(string) (string, bool)) (string, error) {
	value, exists := lookup("STUDY_COURSE")
	if !exists {
		return "Go", nil
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return "", errors.New("STUDY_COURSE는 비어 있을 수 없습니다")
	}
	return value, nil
}

func run() error {
	course, err := courseName(os.LookupEnv)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(os.Stdout, "과정:", course)
	return err
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
