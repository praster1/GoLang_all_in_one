package main

import "fmt"

type stopped struct{ Reason string }

// invoke contains only the known stopped panic used by this demonstration.
func invoke(step func()) (err error) {
	defer func() {
		if value := recover(); value != nil {
			if known, ok := value.(stopped); ok {
				err = fmt.Errorf("step stopped: %s", known.Reason)
				return
			}
			panic(value)
		}
	}()
	step()
	return nil
}
func main() {
	err := invoke(func() { defer fmt.Println("정리 실행"); panic(stopped{Reason: "demonstration"}) })
	fmt.Println("경계 결과:", err)
	fmt.Println("호출자 계속")
}
