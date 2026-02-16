package order

type OrderType int

const (
	ORDER_TYPE_BUY OrderType = iota + 1
	ORDER_TYPE_SELL
)

type OrderStatus int

const (
	ORDER_STATUS_CREATED OrderStatus = iota + 1
	ORDER_STATUS_PENDING
	ORDER_STATUS_COMPLETED
	ORDER_STATUS_REJECTED
)
