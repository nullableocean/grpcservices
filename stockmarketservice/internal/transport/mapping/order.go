package mapping

import (
	stockmarketv1 "github.com/nullableocean/grpcservices/api/gen/stockmarket/v1"
	typesv1 "github.com/nullableocean/grpcservices/api/gen/types/v1"
	"github.com/nullableocean/grpcservices/shared/order"
	"github.com/nullableocean/grpcservices/stockmarketservice/internal/domain"
)

func MapProtoOrderToDomainOrder(pborder *typesv1.Order) *domain.Order {
	return &domain.Order{
		UUID:       pborder.OrderUuid,
		UserUuid:   pborder.UserUuid,
		MarketUuid: pborder.MarketUuid,
		OrderType:  order.OrderType(pborder.Type),
		Price:      MapProtoMoneyToDomain(pborder.Price),
		Quantity:   pborder.Quantity,
	}
}

func MapProtoProcessOrderRequestToDomain(req *stockmarketv1.ProcessOrderRequest) *domain.Order {
	return &domain.Order{
		UUID:       req.Order.OrderUuid,
		UserUuid:   req.Order.UserUuid,
		MarketUuid: req.Order.MarketUuid,
		OrderType:  order.OrderType(req.Order.Type),
		Price:      MapProtoMoneyToDomain(req.Order.Price),
		Quantity:   req.Order.Quantity,
	}
}
