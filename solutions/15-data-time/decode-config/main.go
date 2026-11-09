package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

type Config struct {
	Course       string `json:"course"`
	DailyMinutes int    `json:"daily_minutes"`
}

func readConfig(source io.Reader) (Config, error) {
	const maximum = 4096
	data, err := io.ReadAll(io.LimitReader(source, maximum+1))
	if err != nil {
		return Config{}, fmt.Errorf("설정 읽기: %w", err)
	}
	if len(data) > maximum {
		return Config{}, errors.New("설정은 4096바이트까지 허용합니다")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var config Config
	if err := decoder.Decode(&config); err != nil {
		return Config{}, fmt.Errorf("설정 JSON: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err != nil {
			return Config{}, fmt.Errorf("설정 끝 확인: %w", err)
		}
		return Config{}, errors.New("설정 JSON은 하나만 허용합니다")
	}
	config.Course = strings.TrimSpace(config.Course)
	if config.Course == "" || config.DailyMinutes < 1 || config.DailyMinutes > 120 {
		return Config{}, errors.New("과정과 1~120분이 필요합니다")
	}
	return config, nil
}

func run() error {
	config, err := readConfig(strings.NewReader(`{"course":"Go","daily_minutes":25}`))
	if err != nil {
		return err
	}
	fmt.Printf("%s: 하루 %d분\n", config.Course, config.DailyMinutes)
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
