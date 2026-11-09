package main

func Example_main() {
	main()
	// Output:
	// POST /plans/preview -> 200: {"total_minutes":75}
	// POST /plans/preview -> 200: {"total_minutes":0}
	// POST /plans/preview -> 422: 계획 값을 확인하세요
}
