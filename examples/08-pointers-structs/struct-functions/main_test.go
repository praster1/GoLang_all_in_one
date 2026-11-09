package main

import "testing"

func TestWithdraw(t *testing.T) {
	cases := []struct {
		balance, amount, want int
		ok                    bool
	}{
		{1000, 300, 700, true}, {10, 10, 0, true}, {10, 11, 10, false},
		{10, 0, 10, false}, {10, -1, 10, false}, {-1, 1, -1, false}, {0, 1, 0, false},
	}
	for _, tc := range cases {
		account := Account{Owner: "Mina", Balance: tc.balance}
		ok := withdraw(&account, tc.amount)
		if ok != tc.ok || account.Balance != tc.want || account.Owner != "Mina" {
			t.Fatalf("case %+v: %+v, %t", tc, account, ok)
		}
	}
	if withdraw(nil, 1) {
		t.Fatal("nil accepted")
	}
}

func Example_main() {
	main()
	// Output:
	// 출금: true, 잔액: 700
	// 초과 출금: false, 잔액: 700
	// 계좌 없음: false
}
