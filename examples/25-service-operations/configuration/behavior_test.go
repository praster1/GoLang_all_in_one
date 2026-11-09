package main

import (
	"log/slog"
	"testing"
	"time"
)

func TestConfiguration(t *testing.T) {
	cfg, err := loadConfig(func(string) (string, bool) { return "", false })
	if err != nil || cfg.Address != "127.0.0.1:8080" || cfg.ReadTimeout != 5*time.Second || cfg.Level != slog.LevelInfo {
		t.Fatalf("default=%+v, %v", cfg, err)
	}
	cases := []struct {
		key, value string
		valid      bool
	}{
		{"APP_ADDR", "127.0.0.1:1", true}, {"APP_ADDR", "[::1]:65535", true}, {"APP_ADDR", ":8080", false}, {"APP_ADDR", "localhost:65536", false}, {"APP_ADDR", "localhost:0", false}, {"APP_ADDR", "", false},
		{"APP_READ_TIMEOUT", "1ns", true}, {"APP_READ_TIMEOUT", "1m", true}, {"APP_READ_TIMEOUT", "0s", false}, {"APP_READ_TIMEOUT", "-1s", false}, {"APP_READ_TIMEOUT", "61s", false}, {"APP_READ_TIMEOUT", "abc", false},
		{"APP_LOG_LEVEL", "debug", true}, {"APP_LOG_LEVEL", "warn", true}, {"APP_LOG_LEVEL", "error", true}, {"APP_LOG_LEVEL", "info", true}, {"APP_LOG_LEVEL", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.key+"="+tc.value, func(t *testing.T) {
			_, err := loadConfig(func(key string) (string, bool) {
				if key == tc.key {
					return tc.value, true
				}
				return "", false
			})
			if (err == nil) != tc.valid {
				t.Fatalf("error=%v", err)
			}
		})
	}
}
