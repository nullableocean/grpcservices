package client

import (
	modelsv1 "github.com/nullableocean/grpcservices/api/gen/models/v1"
	"github.com/nullableocean/grpcservices/orderserviceclient/internal/model"
	"github.com/shopspring/decimal"
)

func MapDecimalToProtoMoney(dec decimal.Decimal) *modelsv1.Money {
	units := dec.IntPart()
	nanos := dec.Sub(decimal.NewFromInt(units)).Mul(decimal.NewFromInt(1e9)).IntPart()

	return &modelsv1.Money{
		Units: units,
		Nanos: int32(nanos),
	}
}

func MapDecimalToProtoDeciaml(dec decimal.Decimal) *modelsv1.Decimal {
	units := dec.IntPart()
	nanos := dec.Sub(decimal.NewFromInt(units)).Mul(decimal.NewFromInt(1e9)).IntPart()

	return &modelsv1.Decimal{
		Units: units,
		Nanos: int32(nanos),
	}
}

func MapOrderTypeToProtoType(t model.OrderType) modelsv1.OrderType {
	var pbType modelsv1.OrderType

	switch t {
	case model.OrderTypeLimit:
		pbType = modelsv1.OrderType_ORDER_TYPE_LIMIT
	case model.OrderTypeMarket:
		pbType = modelsv1.OrderType_ORDER_TYPE_MARKET
	case model.OrderTypeStopLoss:
		pbType = modelsv1.OrderType_ORDER_TYPE_STOP_LOSS
	case model.OrderTypeTakeProfit:
		pbType = modelsv1.OrderType_ORDER_TYPE_TAKE_PROFIT
	}

	return pbType
}

func MapOrderSideToProtoSide(s model.OrderSide) modelsv1.OrderSide {
	var pbSide modelsv1.OrderSide

	switch s {
	case model.OrderSideBuy:
		pbSide = modelsv1.OrderSide_ORDER_SIDE_BUY
	case model.OrderSideSell:
		pbSide = modelsv1.OrderSide_ORDER_SIDE_SELL
	}

	return pbSide
}

func MapProtoStatusToStatus(pbstatus modelsv1.OrderStatus) model.OrderStatus {
	var status model.OrderStatus
	switch pbstatus {
	case modelsv1.OrderStatus_ORDER_STATUS_CREATED:
		status = model.OrderStatusCreated
	case modelsv1.OrderStatus_ORDER_STATUS_PENDING:
		status = model.OrderStatusPending
	case modelsv1.OrderStatus_ORDER_STATUS_COMPLETED:
		status = model.OrderStatusCompleted
	case modelsv1.OrderStatus_ORDER_STATUS_CANCELLED:
		status = model.OrderStatusCancelled
	case modelsv1.OrderStatus_ORDER_STATUS_REJECTED:
		status = model.OrderStatusRejected
	}

	return status
}
