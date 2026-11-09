package main

import "fmt"

type Rectangle struct{ Width, Height int }

func (rectangle Rectangle) Area() int      { return rectangle.Width * rectangle.Height }
func (rectangle Rectangle) Perimeter() int { return 2 * (rectangle.Width + rectangle.Height) }

func main() {
	rectangle := Rectangle{Width: 3, Height: 4}
	fmt.Println("넓이:", rectangle.Area())
	fmt.Println("둘레:", rectangle.Perimeter())
	fmt.Println("원본:", rectangle.Width, rectangle.Height)
}
