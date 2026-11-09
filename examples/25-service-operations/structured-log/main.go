package main

import (
	"log/slog"
	"os"
)

func withoutTime(groups []string, attr slog.Attr) slog.Attr {
	if len(groups) == 0 && attr.Key == slog.TimeKey {
		return slog.Attr{}
	}
	return attr
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo, ReplaceAttr: withoutTime,
	})).With("service", "study-api")
	logger.Debug("설정 읽기")
	logger.Info("요청 완료", slog.Group("request", "method", "GET", "status", 200))
	logger.Warn("의존 서비스 지연", "dependency", "catalog", "retryable", true)
}
