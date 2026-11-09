package main

import "fmt"

func main() {
	stock := map[string]int{"pen": 3, "book": 2}
	fmt.Println("pen:", stock["pen"])
	stock["pen"] = 5
	stock["note"] = 4
	fmt.Println("수정한 pen:", stock["pen"])
	fmt.Println("등록 수:", len(stock))
	delete(stock, "book")
	fmt.Println("삭제 뒤 수:", len(stock))
}
