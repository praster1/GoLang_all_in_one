// Package label creates small canonical labels.
package label

import (
	"strings"
	"unicode/utf8"
)

func Normalize(value string) (string, bool) {
	value = strings.TrimSpace(value)
	if !utf8.ValidString(value) || value == "" || utf8.RuneCountInString(value) > 20 || strings.ContainsAny(value, "\r\n") {
		return "", false
	}
	return strings.ToLower(value), true
}
