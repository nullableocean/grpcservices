package model

type OrderSide string

const (
	OrderSideBuy     OrderSide = "buy"
	OrderSideSell    OrderSide = "sell"
	OrderSideUnknown OrderSide = "unknown"
)

func (s OrderSide) IsValid() bool {
	return s == OrderSideBuy || s == OrderSideSell
}
