package exchange

type Balance struct {
	Currency  string
	Available float64
}

type Position struct {
	Symbol string
	Side   string // Buy / Sell / None
	Size   float64
	Entry  float64
}

type Exchange interface {
	GetBalance() (*Balance, error)
	GetPosition(symbol string) (*Position, error)
}
