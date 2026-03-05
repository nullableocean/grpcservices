package server

import (
	stockmarketv1 "github.com/nullableocean/grpcservices/api/gen/stockmarket/v1"
	typesv1 "github.com/nullableocean/grpcservices/api/gen/types/v1"
	"github.com/nullableocean/grpcservices/shared/money"
	"github.com/nullableocean/grpcservices/shared/order"
	"github.com/nullableocean/grpcservices/stockmarketservice/internal/domain"
	"github.com/shopspring/decimal"
)

func mapProtoProcessOrderRequestToDomain(req *stockmarketv1.ProcessOrderRequest) *domain.Order {
	return &domain.Order{
		UUID:       req.Order.OrderUuid,
		UserUuid:   req.Order.UserUuid,
		MarketUuid: req.Order.MarketUuid,
		OrderType:  order.OrderType(req.Order.Type),
		Price:      mapProtoMoneyToDomain(req.Order.Price),
		Quantity:   req.Order.Quantity,
	}
}

func mapProtoMoneyToDomain(pbmoney *typesv1.Money) money.Money {
	return money.Money{
		Decimal: mapProtoMoneyToDecimal(pbmoney),
	}
}

func mapProtoMoneyToDecimal(pbmoney *typesv1.Money) decimal.Decimal {
	units := decimal.NewFromInt(pbmoney.Units)
	nanos := decimal.NewFromInt(int64(pbmoney.Nanos))

	result := units.Add(nanos.Div(decimal.NewFromInt(1e9)))
	return result
}
