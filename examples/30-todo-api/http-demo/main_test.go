package main

func Example_main() {
	main()
	// Output:
	// POST /todos -> 201 {"id":1,"title":"Go 복습","done":false}
	// GET /todos/1 -> 200 {"id":1,"title":"Go 복습","done":false}
	// PUT /todos/1 -> 200 {"id":1,"title":"Go 복습","done":true}
	// GET /todos?limit=2 -> 200 [{"id":1,"title":"Go 복습","done":true}]
	// DELETE /todos/1 -> 204
	// GET /todos/1 -> 404 {"error":"not_found"}
}
