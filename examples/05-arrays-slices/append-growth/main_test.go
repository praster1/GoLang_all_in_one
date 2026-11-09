package main

func Example_main() {
	main()
	// Output:
	// 추가 후: [10 20 30 40]
	// 기존 뷰 확장: [10 20 30 40]
	// 새 저장 공간: [99 20 30 40 50]
	// 기존 값: [10 20]
	// 필요한 용량 확보: true
}
