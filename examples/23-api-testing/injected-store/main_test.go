package main

func Example_main() {
	main()
	// Output:
	// 200 {"id":1,"title":"Go"}
	// 404 {"code":"not_found","message":"학습을 찾을 수 없습니다"}
	// 400 {"code":"invalid_id","message":"ID는 양수 정수입니다"}
}
