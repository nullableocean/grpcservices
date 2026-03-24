package mapping

import (
	modelsv1 "github.com/nullableocean/grpcservices/api/gen/models/v1"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
)

func MapProtoTypeToOrderType(pbType modelsv1.OrderType) model.OrderType {
	var orderType model.OrderType
	switch pbType {
	case modelsv1.OrderType_ORDER_TYPE_LIMIT:
		orderType = model.OrderTypeLimit
	case modelsv1.OrderType_ORDER_TYPE_MARKET:
		orderType = model.OrderTypeMarket
	case modelsv1.OrderType_ORDER_TYPE_STOP_LOSS:
		orderType = model.OrderTypeStopLoss
	case modelsv1.OrderType_ORDER_TYPE_TAKE_PROFIT:
		orderType = model.OrderTypeTakeProfit
	}

	return orderType
}

func MapProtoSideToOrderSide(pbSide modelsv1.OrderSide) model.OrderSide {
	var side model.OrderSide
	switch pbSide {
	case modelsv1.OrderSide_ORDER_SIDE_BUY:
		side = model.OrderSideBuy
	case modelsv1.OrderSide_ORDER_SIDE_SELL:
		side = model.OrderSideSell
	}

	return side
}

func MapOrderStatusToProtoStatus(s model.OrderStatus) modelsv1.OrderStatus {
	var pbStatus modelsv1.OrderStatus
	switch s {
	case model.OrderStatusCreated:
		pbStatus = modelsv1.OrderStatus_ORDER_STATUS_CREATED
	case model.OrderStatusPending:
		pbStatus = modelsv1.OrderStatus_ORDER_STATUS_PENDING
	case model.OrderStatusCompleted:
		pbStatus = modelsv1.OrderStatus_ORDER_STATUS_COMPLETED
	case model.OrderStatusCancelled:
		pbStatus = modelsv1.OrderStatus_ORDER_STATUS_CANCELLED
	case model.OrderStatusRejected:
		pbStatus = modelsv1.OrderStatus_ORDER_STATUS_REJECTED
	}

	return pbStatus
}
