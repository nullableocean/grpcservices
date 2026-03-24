package model

type OrderType string

const (
	OrderTypeLimit      OrderType = "limit"
	OrderTypeMarket     OrderType = "market"
	OrderTypeStopLoss   OrderType = "stop"
	OrderTypeTakeProfit OrderType = "profit"
)

func (t OrderType) IsValid() bool {
	switch t {
	case OrderTypeLimit, OrderTypeMarket, OrderTypeStopLoss, OrderTypeTakeProfit:
		return true
	}
	return false
}
