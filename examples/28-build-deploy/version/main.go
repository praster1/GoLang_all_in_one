package main

import "fmt"

var version = "dev"
var revision = "unknown"

func versionLine(v, r string) string { return fmt.Sprintf("version=%s revision=%s", v, r) }

func main() { fmt.Println(versionLine(version, revision)) }
