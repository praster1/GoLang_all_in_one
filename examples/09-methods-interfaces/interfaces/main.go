package main

import "fmt"

type Describer interface{ Describe() string }

type Book struct{ Title string }

func (book Book) Describe() string { return "책: " + book.Title }

type User struct{ Name string }

func (user User) Describe() string { return "사용자: " + user.Name }

func printDescription(value Describer) {
	fmt.Println(value.Describe())
}

func main() {
	values := []Describer{Book{Title: "Go"}, User{Name: "Mina"}}
	for _, value := range values {
		printDescription(value)
	}
}
