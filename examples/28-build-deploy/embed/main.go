package main

import (
	_ "embed"
	"fmt"
	"strings"
)

//go:embed message.txt
var message string

func greeting() string { return strings.TrimSpace(message) }

func main() { fmt.Println(greeting()) }
