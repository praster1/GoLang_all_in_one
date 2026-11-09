package main

import "fmt"

type InputError struct{ Message string }

func (e *InputError) Error() string { return e.Message }
func wrong() error                  { var err *InputError; return err }
func correct() error                { return nil }
func main() {
	fmt.Println("typed nil:", wrong() == nil)
	fmt.Println("진짜 nil:", correct() == nil)
}
