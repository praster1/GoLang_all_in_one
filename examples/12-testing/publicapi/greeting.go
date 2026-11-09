package publicapi

import "strings"

// Greeting returns a greeting, using 방문자 for an empty or whitespace-only name.
func Greeting(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "방문자"
	}
	return name + "님, 안녕하세요."
}
