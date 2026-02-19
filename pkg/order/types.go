package order

type OrderType int

const (
	ORDER_TYPE_BUY OrderType = iota + 1
	ORDER_TYPE_SELL
)

func MapOrderTypeToString(t OrderType) string {
	switch t {
	case ORDER_TYPE_BUY:
		return "buy"
	case ORDER_TYPE_SELL:
		return "sell"
	}

	return ""
}

type OrderStatus int

const (
	ORDER_STATUS_CREATED OrderStatus = iota + 1
	ORDER_STATUS_PENDING
	ORDER_STATUS_COMPLETED
	ORDER_STATUS_REJECTED
)

func MapOrderStatusToString(t OrderStatus) string {
	switch t {
	case ORDER_STATUS_CREATED:
		return "created"
	case ORDER_STATUS_PENDING:
		return "pending"
	case ORDER_STATUS_COMPLETED:
		return "completed"
	case ORDER_STATUS_REJECTED:
		return "rejected"
	}

	return ""
}
