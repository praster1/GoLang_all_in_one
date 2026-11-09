package main

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestSecretRedaction(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	secret := "never-publish-this-test-value"
	logger.Info("configuration", "secret", Secret(secret), "credentials", Credentials{Client: "worker", Token: Secret(secret)})
	if strings.Contains(output.String(), secret) {
		t.Fatal("secret leaked")
	}
	if strings.Count(output.String(), "[REDACTED]") != 2 || !strings.Contains(output.String(), "worker") {
		t.Fatalf("record=%s", &output)
	}
}
