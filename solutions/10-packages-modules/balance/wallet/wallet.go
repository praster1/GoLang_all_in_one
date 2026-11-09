// Package wallet provides a small integer balance with explicit limits.
package wallet

type Wallet struct{ balance int }

func New(initial int) (Wallet, bool) {
	if initial < 0 || initial > 10000 {
		return Wallet{}, false
	}
	return Wallet{balance: initial}, true
}
func (w Wallet) Balance() int { return w.balance }
func (w *Wallet) Spend(amount int) bool {
	if w == nil || amount < 0 || amount > w.balance {
		return false
	}
	w.balance -= amount
	return true
}
