// Package counter keeps a count in the range 0..100.
package counter

// Counter has a useful zero value. The count is changed through Add.
type Counter struct{ value int }

func (c Counter) Value() int { return c.value }
func (c *Counter) Add(amount int) bool {
	if c == nil || amount < 0 || amount > 100-c.value {
		return false
	}
	c.value += amount
	return true
}
