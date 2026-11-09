package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Address     string
	ReadTimeout time.Duration
	Level       slog.Level
}

func loadConfig(lookup func(string) (string, bool)) (Config, error) {
	cfg := Config{Address: "127.0.0.1:8080", ReadTimeout: 5 * time.Second, Level: slog.LevelInfo}
	if value, ok := lookup("APP_ADDR"); ok {
		host, port, err := net.SplitHostPort(value)
		if err != nil || host == "" {
			return Config{}, errors.New("APP_ADDR는 호스트:포트 형식이어야 합니다")
		}
		n, err := strconv.Atoi(port)
		if err != nil || n < 1 || n > 65535 {
			return Config{}, errors.New("APP_ADDR 포트는 1~65535여야 합니다")
		}
		cfg.Address = value
	}
	if value, ok := lookup("APP_READ_TIMEOUT"); ok {
		d, err := time.ParseDuration(value)
		if err != nil || d <= 0 || d > time.Minute {
			return Config{}, errors.New("APP_READ_TIMEOUT은 0보다 크고 1m 이하여야 합니다")
		}
		cfg.ReadTimeout = d
	}
	if value, ok := lookup("APP_LOG_LEVEL"); ok {
		switch value {
		case "debug":
			cfg.Level = slog.LevelDebug
		case "info":
			cfg.Level = slog.LevelInfo
		case "warn":
			cfg.Level = slog.LevelWarn
		case "error":
			cfg.Level = slog.LevelError
		default:
			return Config{}, errors.New("APP_LOG_LEVEL은 debug/info/warn/error 중 하나여야 합니다")
		}
	}
	return cfg, nil
}

func run() error {
	env := map[string]string{"APP_ADDR": "127.0.0.1:9090", "APP_READ_TIMEOUT": "2s", "APP_LOG_LEVEL": "warn"}
	cfg, err := loadConfig(func(key string) (string, bool) { value, ok := env[key]; return value, ok })
	if err != nil {
		return err
	}
	fmt.Println("주소:", cfg.Address)
	fmt.Println("읽기 제한:", cfg.ReadTimeout)
	fmt.Println("로그 수준:", cfg.Level)
	return nil
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
