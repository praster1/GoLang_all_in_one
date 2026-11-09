package main

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"testing"
)

func TestStructuredLog(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, &slog.HandlerOptions{Level: slog.LevelWarn, ReplaceAttr: withoutTime})).With("service", "demo")
	logger.Info("hidden")
	logger.Warn("problem", "attempt", 2)
	var item map[string]any
	if err := json.Unmarshal(output.Bytes(), &item); err != nil {
		t.Fatal(err)
	}
	if item["msg"] != "problem" || item["level"] != "WARN" || item["attempt"] != float64(2) || item["service"] != "demo" {
		t.Fatalf("record=%v", item)
	}
	if _, ok := item["time"]; ok {
		t.Fatal("unexpected clock")
	}
}
