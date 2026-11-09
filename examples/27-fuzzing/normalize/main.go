package main

import (
	"fmt"
	"strings"
)

func normalizeSpace(text string) string { return strings.Join(strings.Fields(text), " ") }

func main() { fmt.Println(normalizeSpace("  Go\t 공부\n 시작  ")) }
