package main

func Example_main() {
	main()
	// Output:
	// POST /lessons -> 201: {"id":1,"title":"Go","minutes":25}
	// POST /lessons -> 422: 학습 값의 범위를 확인하세요
	// POST /lessons -> 400: 잘못된 JSON입니다
}
