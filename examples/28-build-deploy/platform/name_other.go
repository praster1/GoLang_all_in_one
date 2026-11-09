//go:build !linux && !windows

package main

func platformName() string { return "other" }
