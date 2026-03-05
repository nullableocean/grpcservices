package mapping

import (
	orderv1 "github.com/nullableocean/grpcservices/api/gen/order/v1"
	typesv1 "github.com/nullableocean/grpcservices/api/gen/types/v1"
	"github.com/nullableocean/grpcservices/orderservice/internal/domain"
	"github.com/nullableocean/grpcservices/orderservice/internal/dto"
	"github.com/nullableocean/grpcservices/shared/order"
)

// Map proto request to service order dto
func MapCreateOrderRequestToOrderDto(req *orderv1.CreateOrderRequest) *dto.CreateOrderDto {
	return &dto.CreateOrderDto{
		UserUuid:   req.UserUuid,
		MarketUuid: req.MarketId,
		Price:      MapProtoMoneyToDomain(req.Price),
		Quantity:   req.Quantity,
		OrderType:  order.OrderType(req.OrderType),
	}
}

// Map order to create order response
func MapDomainOrderToProtoResponse(order *domain.Order) *orderv1.CreateOrderResponse {
	return &orderv1.CreateOrderResponse{
		OrderUuid: order.Id(),
		Status:    typesv1.OrderStatus(order.GetStatus()),
	}
}
