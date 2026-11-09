package main

func Example_main() {
	main()
	// Output:
	// GET /health -> 200: ready
	// POST /health -> 405: Method Not Allowed
}
