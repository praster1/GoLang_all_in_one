package main

import (
	"fmt"
	"runtime"
)

func main() { fmt.Printf("target=%s/%s branch=%s\n", runtime.GOOS, runtime.GOARCH, platformName()) }
