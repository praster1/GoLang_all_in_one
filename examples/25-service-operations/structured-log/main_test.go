package main

func Example_main() {
	main()
	// Output:
	// {"level":"INFO","msg":"요청 완료","service":"study-api","request":{"method":"GET","status":200}}
	// {"level":"WARN","msg":"의존 서비스 지연","service":"study-api","dependency":"catalog","retryable":true}
}
