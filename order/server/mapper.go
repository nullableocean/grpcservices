package server

import (
	"github.com/nullableocean/grpcservices/api/orderpb"
	"github.com/nullableocean/grpcservices/order/domain"
	"github.com/nullableocean/grpcservices/pkg/order"
)

type OrderServerMapper struct{}

func (mapper *OrderServerMapper) CreateOrderRequestToOrderDto(req *orderpb.CreateOrderRequest) domain.CreateOrderDto {
	return domain.CreateOrderDto{
		MarketId:  req.MarketId,
		Price:     float64(req.Price),
		Quantity:  req.Quantity,
		OrderType: order.OrderType(req.OrderType),
	}
}

func (mapper *OrderServerMapper) OrderToPbOrderResponse(order *domain.Order) *orderpb.CreateOrderResponse {
	return &orderpb.CreateOrderResponse{
		OrderId: order.Id(),
		Status:  orderpb.OrderStatus(order.Status()),
	}
}
