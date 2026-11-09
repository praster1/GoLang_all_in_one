package main

import "fmt"

type Account struct {
	Owner   string
	Balance int
}

func withdraw(account *Account, amount int) bool {
	if account == nil || account.Balance < 0 || amount <= 0 || amount > account.Balance {
		return false
	}
	account.Balance -= amount
	return true
}

func main() {
	account := Account{Owner: "Mina", Balance: 1000}
	ok := withdraw(&account, 300)
	fmt.Printf("출금: %t, 잔액: %d\n", ok, account.Balance)
	ok = withdraw(&account, 800)
	fmt.Printf("초과 출금: %t, 잔액: %d\n", ok, account.Balance)
	fmt.Println("계좌 없음:", withdraw(nil, 100))
}
