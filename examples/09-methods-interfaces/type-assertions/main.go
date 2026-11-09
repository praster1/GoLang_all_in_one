package main

import "fmt"

func describeValue(value any) string {
	switch typed := value.(type) {
	case nil:
		return "값 없음"
	case string:
		return fmt.Sprintf("문자열 %q", typed)
	case int:
		return fmt.Sprintf("정수 %d", typed)
	default:
		return fmt.Sprintf("다른 타입 %T", typed)
	}
}

func main() {
	var value any = 42
	number, ok := value.(int)
	fmt.Println("int 확인:", number, ok)
	text, ok := value.(string)
	fmt.Printf("string 확인: %q %t\n", text, ok)
	for _, item := range []any{"Go", 7, true, nil} {
		fmt.Println(describeValue(item))
	}
}
