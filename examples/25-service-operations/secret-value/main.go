package main

import (
	"log/slog"
	"os"
)

type Secret string

func (Secret) LogValue() slog.Value { return slog.StringValue("[REDACTED]") }

type Credentials struct {
	Client string
	Token  Secret
}

func (c Credentials) LogValue() slog.Value {
	return slog.GroupValue(slog.String("client", c.Client), slog.Any("token", c.Token))
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if len(groups) == 0 && a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	}))
	credentials := Credentials{Client: "lesson-client", Token: Secret("classroom-placeholder")}
	logger.Info("클라이언트 설정", "credentials", credentials)
}
