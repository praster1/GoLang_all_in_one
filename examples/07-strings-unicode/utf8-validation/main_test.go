package main

func Example_main() {
	main()
	// Output:
	// 유효한 UTF-8: false
	// 0: U+0041
	// 1: U+FFFD
	// 2: U+0042
	// 잘못된 바이트: U+FFFD, 크기: 1
	// 실제 대체 문자: U+FFFD, 크기: 3
}
