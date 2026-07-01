package client

import (
	modelsv1 "github.com/nullableocean/grpcservices/api/gen/models/v1"
	"github.com/nullableocean/grpcservices/orderserviceclient/internal/model"
	"github.com/shopspring/decimal"
)

func MapProtoMoneyToDecimal(pbMoney *modelsv1.Money) decimal.Decimal {
	if pbMoney == nil {
		return decimal.Zero
	}

	return decimal.New(pbMoney.Units, 0).Add(decimal.New(int64(pbMoney.Nanos), -9))
}

func MapProtoDecimalToDecimal(pbDecimal *modelsv1.Decimal) decimal.Decimal {
	if pbDecimal == nil {
		return decimal.Zero
	}

	return decimal.New(pbDecimal.Units, 0).Add(decimal.New(int64(pbDecimal.Nanos), -9))
}

func MapDecimalToProtoMoney(dec decimal.Decimal) *modelsv1.Money {
	units := dec.IntPart()
	nanos := dec.Sub(decimal.NewFromInt(units)).Mul(decimal.NewFromInt(1e9)).IntPart()

	return &modelsv1.Money{
		Units: units,
		Nanos: int32(nanos),
	}
}

func MapDecimalToProtoDecimal(dec decimal.Decimal) *modelsv1.Decimal {
	units := dec.IntPart()
	nanos := dec.Sub(decimal.NewFromInt(units)).Mul(decimal.NewFromInt(1e9)).IntPart()

	return &modelsv1.Decimal{
		Units: units,
		Nanos: int32(nanos),
	}
}

func MapOrderTypeToProtoType(t model.OrderType) modelsv1.OrderType {
	switch t {
	case model.OrderTypeLimit:
		return modelsv1.OrderType_ORDER_TYPE_LIMIT
	case model.OrderTypeMarket:
		return modelsv1.OrderType_ORDER_TYPE_MARKET
	case model.OrderTypeStopLoss:
		return modelsv1.OrderType_ORDER_TYPE_STOP_LOSS
	case model.OrderTypeTakeProfit:
		return modelsv1.OrderType_ORDER_TYPE_TAKE_PROFIT
	default:
		return modelsv1.OrderType_ORDER_TYPE_UNSPECIFIED
	}
}

func MapProtoOrderTypeToType(pbType modelsv1.OrderType) model.OrderType {
	switch pbType {
	case modelsv1.OrderType_ORDER_TYPE_LIMIT:
		return model.OrderTypeLimit
	case modelsv1.OrderType_ORDER_TYPE_MARKET:
		return model.OrderTypeMarket
	case modelsv1.OrderType_ORDER_TYPE_STOP_LOSS:
		return model.OrderTypeStopLoss
	case modelsv1.OrderType_ORDER_TYPE_TAKE_PROFIT:
		return model.OrderTypeTakeProfit
	default:
		return model.OrderTypeUnknown
	}
}

func MapOrderSideToProtoSide(s model.OrderSide) modelsv1.OrderSide {
	switch s {
	case model.OrderSideBuy:
		return modelsv1.OrderSide_ORDER_SIDE_BUY
	case model.OrderSideSell:
		return modelsv1.OrderSide_ORDER_SIDE_SELL
	default:
		return modelsv1.OrderSide_ORDER_SIDE_UNSPECIFIED
	}
}

func MapProtoOrderSideToSide(pbSide modelsv1.OrderSide) model.OrderSide {
	switch pbSide {
	case modelsv1.OrderSide_ORDER_SIDE_BUY:
		return model.OrderSideBuy
	case modelsv1.OrderSide_ORDER_SIDE_SELL:
		return model.OrderSideSell
	default:
		return model.OrderSideUnknown
	}
}

func MapStatusToProtoStatus(status model.OrderStatus) modelsv1.OrderStatus {
	switch status {
	case model.OrderStatusCreated:
		return modelsv1.OrderStatus_ORDER_STATUS_CREATED
	case model.OrderStatusPending:
		return modelsv1.OrderStatus_ORDER_STATUS_PENDING
	case model.OrderStatusCompleted:
		return modelsv1.OrderStatus_ORDER_STATUS_COMPLETED
	case model.OrderStatusCancelled:
		return modelsv1.OrderStatus_ORDER_STATUS_CANCELLED
	case model.OrderStatusRejected:
		return modelsv1.OrderStatus_ORDER_STATUS_REJECTED
	default:
		return modelsv1.OrderStatus_ORDER_STATUS_UNSPECIFIED
	}
}

func MapProtoStatusToStatus(pbstatus modelsv1.OrderStatus) model.OrderStatus {
	switch pbstatus {
	case modelsv1.OrderStatus_ORDER_STATUS_CREATED:
		return model.OrderStatusCreated
	case modelsv1.OrderStatus_ORDER_STATUS_PENDING:
		return model.OrderStatusPending
	case modelsv1.OrderStatus_ORDER_STATUS_COMPLETED:
		return model.OrderStatusCompleted
	case modelsv1.OrderStatus_ORDER_STATUS_CANCELLED:
		return model.OrderStatusCancelled
	case modelsv1.OrderStatus_ORDER_STATUS_REJECTED:
		return model.OrderStatusRejected
	default:
		return model.OrderStatusUnknown
	}
}

func MapProtoOrderToOrder(pborder *modelsv1.Order) *model.Order {
	if pborder == nil {
		return nil
	}

	return &model.Order{
		UUID:       pborder.OrderUuid,
		MarketUUID: pborder.UserUuid,
		Type:       MapProtoOrderTypeToType(pborder.Type),
		Status:     MapProtoStatusToStatus(pborder.Status),
		Side:       MapProtoOrderSideToSide(pborder.Side),
		Price:      MapProtoMoneyToDecimal(pborder.Price),
		Quantity:   MapProtoDecimalToDecimal(pborder.Quantity),
		CreatedAt:  pborder.CreatedAt.AsTime(),
		UpdatedAt:  pborder.UpdatedAt.AsTime(),
	}
}
