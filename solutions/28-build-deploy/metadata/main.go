package main

import (
	"errors"
	"fmt"
	"regexp"
)

type release struct{ Name, Version, OS, Arch string }

var validLabel = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)

func artifactName(r release) (string, error) {
	if !validLabel.MatchString(r.Name) || !validLabel.MatchString(r.Version) {
		return "", errors.New("invalid release label")
	}
	if r.OS != "linux" && r.OS != "windows" && r.OS != "darwin" {
		return "", errors.New("unsupported OS")
	}
	if r.Arch != "amd64" && r.Arch != "arm64" {
		return "", errors.New("unsupported architecture")
	}
	suffix := ""
	if r.OS == "windows" {
		suffix = ".exe"
	}
	return fmt.Sprintf("%s-%s-%s-%s%s", r.Name, r.Version, r.OS, r.Arch, suffix), nil
}

func main() {
	name, err := artifactName(release{"hello", "1.0.0", "windows", "amd64"})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(name)
}
