package domain

import (
	"github.com/nullableocean/grpcservices/shared/money"
	"github.com/nullableocean/grpcservices/shared/order"
)

type Order struct {
	UUID       string
	UserUuid   string
	MarketUuid string
	OrderType  order.OrderType
	Price      money.Money
	Quantity   int64
}

func (o *Order) IsBuy() bool {
	return o.OrderType == order.ORDER_TYPE_BUY
}

func (o *Order) IsSell() bool {
	return o.OrderType == order.ORDER_TYPE_SELL
}
