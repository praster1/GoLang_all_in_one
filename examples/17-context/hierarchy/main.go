package main

import (
	"context"
	"fmt"
)

func main() {
	parent, cancelParent := context.WithCancel(context.Background())
	defer cancelParent()
	child, cancelChild := context.WithCancel(parent)
	defer cancelChild()
	sibling, cancelSibling := context.WithCancel(parent)
	defer cancelSibling()

	cancelChild()
	<-child.Done()
	fmt.Println("자식 취소:", child.Err() != nil)
	fmt.Println("부모 유지:", parent.Err() == nil)
	fmt.Println("형제 유지:", sibling.Err() == nil)

	cancelParent()
	<-sibling.Done()
	fmt.Println("부모 취소 후 형제:", sibling.Err() != nil)
}
