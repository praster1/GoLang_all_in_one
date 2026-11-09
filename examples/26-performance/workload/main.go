package main

import "fmt"

func checksum(data []byte, rounds int) uint64 {
	var sum uint64
	for r := 0; r < rounds; r++ {
		for _, value := range data {
			sum = sum*33 + uint64(value)
		}
	}
	return sum
}

func main() {
	fmt.Println("체크섬:", checksum([]byte("Go"), 2))
}
