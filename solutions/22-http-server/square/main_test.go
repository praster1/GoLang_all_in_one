package main

func Example_main() {
	main()
	// Output:
	// GET /squares/12 -> 200: {"input":12,"square":144}
	// GET /squares/0 -> 200: {"input":0,"square":0}
	// GET /squares/1001 -> 400: {"error":"invalid_number"}
}
