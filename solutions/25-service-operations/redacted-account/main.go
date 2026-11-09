package main

import (
	"log/slog"
	"os"
)

type Account struct{ ID, Email, Password string }

func (a Account) LogValue() slog.Value { return slog.GroupValue(slog.String("id", a.ID)) }
func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
		if len(groups) == 0 && a.Key == slog.TimeKey {
			return slog.Attr{}
		}
		return a
	}}))
	logger.Info("account", "account", Account{ID: "student-7", Email: "demo@example.invalid", Password: "placeholder"})
}
