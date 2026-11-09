package main

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

func TestAllowList(t *testing.T) {
	var out bytes.Buffer
	l := slog.New(slog.NewJSONHandler(&out, nil))
	l.Info("account", "account", Account{"id-1", "private@example.invalid", "private-pass"})
	if strings.Contains(out.String(), "private") {
		t.Fatal("private value leaked")
	}
	var got struct {
		Account map[string]string `json:"account"`
	}
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Account) != 1 || got.Account["id"] != "id-1" {
		t.Fatalf("fields=%v", got.Account)
	}
}
