package main

import "fmt"

type Book struct {
	Title     string
	Price     int
	Available bool
}

func main() {
	var empty Book
	fmt.Printf("제로 값: %q, %d, %t\n", empty.Title, empty.Price, empty.Available)
	book := Book{Title: "Go 첫걸음", Price: 12000, Available: true}
	copied := book
	copied.Price = 8000
	fmt.Printf("원본: %s, %d원\n", book.Title, book.Price)
	fmt.Printf("복사본: %+v\n", copied)
	fmt.Println("같은 구조체 값:", book == copied)
	pointer := &book
	pointer.Available = false
	fmt.Println("판매 가능:", book.Available)
}
