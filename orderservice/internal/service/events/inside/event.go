package inside

import (
	"time"

	"github.com/nullableocean/grpcservices/orderservice/internal/domain"
	"github.com/nullableocean/grpcservices/shared/order"
)

type Event interface {
	EventType() string
}

type EventType string

const (
	EVENT_NEW_ORDER_STATUS EventType = "new_status"
	EVENT_CREATED_ORDER    EventType = "created_order"
)

type NewStatusEvent struct {
	OrderUuid string
	NewStatus order.OrderStatus
	UpdatedAt time.Time
}

func (e *NewStatusEvent) EventType() string {
	return string(EVENT_NEW_ORDER_STATUS)
}

type OrderCreatedEvent struct {
	Order *domain.Order
}

func (e *OrderCreatedEvent) EventType() string {
	return string(EVENT_CREATED_ORDER)
}
