// Package lesson defines the shared input policy of two small commands.
package lesson

import (
	"fmt"
	"strings"
)

type Lesson struct {
	title   string
	minutes int
}

func New(title string, minutes int) (Lesson, bool) {
	title = strings.TrimSpace(title)
	if title == "" || minutes < 1 || minutes > 240 {
		return Lesson{}, false
	}
	return Lesson{title: title, minutes: minutes}, true
}
func (value Lesson) Summary() string { return fmt.Sprintf("%s: %d분", value.title, value.minutes) }
