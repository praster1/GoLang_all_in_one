// Package greet formats short greetings without producing output.
package greet

import "strings"

// Hello returns a greeting after trimming surrounding spaces.
func Hello(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "방문자"
	}
	return "안녕하세요, " + name + "!"
}
