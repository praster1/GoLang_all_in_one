package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"strings"
)

func digest(reader io.Reader) (string, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, reader); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func main() {
	sum, err := digest(strings.NewReader("abc"))
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(sum)
}
