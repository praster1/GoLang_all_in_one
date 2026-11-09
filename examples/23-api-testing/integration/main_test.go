package main

func Example_main() {
	main()
	// Output:
	// GET 200 {"id":1,"title":"Go"}
	// HEAD 200 (빈 본문)
	// GET 404 {"code":"not_found","message":"학습을 찾을 수 없습니다"}
}
