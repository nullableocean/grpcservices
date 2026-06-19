package mapping

import (
	modelsv1 "github.com/nullableocean/grpcservices/api/gen/models/v1"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func MapProtoTypeToOrderType(pbType modelsv1.OrderType) model.OrderType {
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
		return model.UndefinedOrderType
	}

}

func MapProtoSideToOrderSide(pbSide modelsv1.OrderSide) model.OrderSide {
	switch pbSide {
	case modelsv1.OrderSide_ORDER_SIDE_BUY:
		return model.OrderSideBuy

	case modelsv1.OrderSide_ORDER_SIDE_SELL:
		return model.OrderSideSell
	default:
		return model.UndefinedOrderSide
	}
}

func MapOrderTypeToProto(orderType model.OrderType) modelsv1.OrderType {
	switch orderType {
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

func MapOrderSideToProto(orderSide model.OrderSide) modelsv1.OrderSide {
	switch orderSide {
	case model.OrderSideBuy:
		return modelsv1.OrderSide_ORDER_SIDE_BUY
	case model.OrderSideSell:
		return modelsv1.OrderSide_ORDER_SIDE_SELL
	default:
		return modelsv1.OrderSide_ORDER_SIDE_UNSPECIFIED
	}
}

func MapOrderStatusToProtoStatus(s model.OrderStatus) modelsv1.OrderStatus {
	switch s {
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

func MapProtoStatusToOrderStatus(pbStatus modelsv1.OrderStatus) model.OrderStatus {
	switch pbStatus {
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
		return model.OrderStatusUndefined
	}
}

func MapOrderToProtoOrder(o *model.Order) *modelsv1.Order {
	return &modelsv1.Order{
		OrderUuid:  o.UUID,
		UserUuid:   o.UserUUID,
		MarketUuid: o.MarketUUID,
		Type:       MapOrderTypeToProto(o.Type),
		Status:     MapOrderStatusToProtoStatus(o.Status),
		Side:       MapOrderSideToProto(o.Side),
		Price:      MapDecimalToProtoMoney(o.Price),
		Quantity:   MapDecimalToProtoDecimal(o.Quantity),
		CreatedAt:  timestamppb.New(o.CreatedAt),
		UpdatedAt:  timestamppb.New(o.UpdatedAt),
	}
}
