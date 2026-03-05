package order

type OrderType int

const (
	ORDER_TYPE_BUY OrderType = iota + 1
	ORDER_TYPE_SELL
)

func (t OrderType) String() string {
	switch t {
	case ORDER_TYPE_BUY:
		return "buy"
	case ORDER_TYPE_SELL:
		return "sell"
	}

	return ""
}
