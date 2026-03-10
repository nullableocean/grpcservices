package domain

import (
	"time"

	"github.com/nullableocean/grpcservices/shared/money"
	"github.com/nullableocean/grpcservices/shared/order"
)

type Order struct {
	UUID       string
	UserUuid   string
	MarketUuid string
	Price      money.Money
	Quantity   int64
	Status     order.OrderStatus
	OrderType  order.OrderType
	CreatedAt  time.Time
}

func (o *Order) Id() string {
	return o.UUID
}

func (o *Order) GetUserUuid() string {
	return o.UserUuid
}

func (o *Order) MarketId() string {
	return o.MarketUuid
}

func (o *Order) GetPrice() money.Money {
	return o.Price
}

func (o *Order) GetQuantity() int64 {
	return o.Quantity
}

func (o *Order) GetType() order.OrderType {
	return o.OrderType
}

func (o *Order) GetStatus() order.OrderStatus {
	return o.Status
}
