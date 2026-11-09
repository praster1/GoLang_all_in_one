package main

import (
	"encoding/hex"
	"fmt"
)

func encode(data []byte) string { return hex.EncodeToString(data) }

func main() {
	data := []byte{0, 15, 255}
	text := encode(data)
	decoded, err := hex.DecodeString(text)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(text)
	fmt.Println(decoded)
}
