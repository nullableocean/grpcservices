package mapping

import (
	modelsv1 "github.com/nullableocean/grpcservices/api/gen/models/v1"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
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
