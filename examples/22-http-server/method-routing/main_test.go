package main

func Example_main() {
	main()
	// Output:
	// GET / -> 200: 목록
	// POST /lessons -> 201: 생성됨
	// GET /lessons/7 -> 200: 학습 7
	// DELETE /lessons/7 -> 405: Method Not Allowed
	// GET /missing -> 404: 404 page not found
}
