package main

func Example_main() {
	main()
	// Output:
	// 200 {"title":"Go","minutes":25}
	// 422 {"code":"missing_field","message":"title과 minutes가 필요합니다"}
	// 422 {"code":"invalid_minutes","message":"분량은 1~180입니다"}
}
