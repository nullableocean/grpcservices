package domain

import "github.com/nullableocean/grpcservices/pkg/order"

type Order struct {
	id        int64
	userId    int64
	marketId  int64
	price     float64
	quantity  int64
	status    order.OrderStatus
	orderType order.OrderType
}

type CreateOrderDto struct {
	MarketId  int64
	Price     float64
	Quantity  int64
	OrderType order.OrderType
}

func NewOrder(id int64, userId int64, data CreateOrderDto) *Order {
	return &Order{
		id:        id,
		userId:    userId,
		marketId:  data.MarketId,
		price:     data.Price,
		quantity:  data.Quantity,
		orderType: data.OrderType,
		status:    order.ORDER_STATUS_CREATED,
	}
}

func (o *Order) Id() int64 {
	return o.id
}

func (o *Order) UserId() int64 {
	return o.userId
}

func (o *Order) MarketId() int64 {
	return o.marketId
}

func (o *Order) Price() float64 {
	return o.price
}

func (o *Order) Quantity() int64 {
	return o.quantity
}

func (o *Order) SetStatus(status order.OrderStatus) {
	o.status = status
}

func (o *Order) Status() order.OrderStatus {
	return o.status
}

func (o *Order) OrderType() order.OrderType {
	return o.orderType
}
