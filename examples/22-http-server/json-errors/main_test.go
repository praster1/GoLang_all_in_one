package main

func Example_main() {
	main()
	// Output:
	// GET /lessons/x -> 400: {"error":"invalid_id"}
	// GET /lessons/2 -> 404: {"error":"lesson_not_found"}
	// GET /lessons/1 -> 200: {"id":1,"title":"Go"}
}
