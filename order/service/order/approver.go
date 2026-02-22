package order

import (
	"context"
	"fmt"
	"slices"

	"github.com/nullableocean/grpcservices/order/domain"
	"github.com/nullableocean/grpcservices/order/service"
	"github.com/nullableocean/grpcservices/pkg/order"
)

var (
	ErrStatusUnavailable = fmt.Errorf("%w: status unavailable for order", service.ErrInvalidData)
)

type StatusApprover struct {
}

func (sa *StatusApprover) CanChangeStatus(ctx context.Context, order *domain.Order, newStatus order.OrderStatus) error {
	currentStatus := order.Status()
	allowedStatuses := sa.allowedNewStatuses(currentStatus)

	if slices.Contains(allowedStatuses, newStatus) {
		return nil
	}

	return ErrStatusUnavailable
}

func (sa *StatusApprover) allowedNewStatuses(current order.OrderStatus) []order.OrderStatus {
	switch current {
	case order.ORDER_STATUS_CREATED:
		return []order.OrderStatus{order.ORDER_STATUS_PENDING, order.ORDER_STATUS_COMPLETED, order.ORDER_STATUS_REJECTED}
	case order.ORDER_STATUS_PENDING:
		return []order.OrderStatus{order.ORDER_STATUS_COMPLETED, order.ORDER_STATUS_REJECTED}
	case order.ORDER_STATUS_COMPLETED:
		return nil
	case order.ORDER_STATUS_REJECTED:
		return nil

	default:
		return nil
	}
}
