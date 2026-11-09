package main

import (
	"runtime"
	"testing"
)

func TestPlatformName(t *testing.T) {
	want := "other"
	if runtime.GOOS == "linux" || runtime.GOOS == "windows" {
		want = runtime.GOOS
	}
	if runtime.GOOS == "android" {
		want = "linux"
	}
	if platformName() != want {
		t.Fatalf("got=%s want=%s", platformName(), want)
	}
}
