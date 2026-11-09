package main

func Example_main() {
	main()
	// Output:
	// GET /lessons/7 -> 200: {"id":7,"title":"Go"}
	// GET /lessons/0 -> 400: ID는 양수 정수여야 합니다
	// GET /lessons/x -> 400: ID는 양수 정수여야 합니다
	// GET /lessons/7/extra -> 404: 404 page not found
}
