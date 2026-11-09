package main

func Example_main() {
	main()
	// Output:
	// 200 {"id":1,"title":"Go"}
	// 404 {"code":"not_found","message":"학습을 찾을 수 없습니다"}
	// 500 {"code":"internal_error","message":"요청을 처리하지 못했습니다"}
	// 400 {"code":"invalid_id","message":"ID는 양수 정수입니다"}
}
