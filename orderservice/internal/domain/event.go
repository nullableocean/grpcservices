package domain

import (
	"time"

	"github.com/nullableocean/grpcservices/shared/order"
)

type EventStatus int

const (
	EVENT_STATUS_CREATED EventStatus = iota
	EVENT_STATUS_PROCESSING
	EVENT_STATUS_PROCESSED
	EVENT_STATUS_ERROR
)

type UpdateEvent struct {
	UUID      string
	OrderUuid string
	NewStatus order.OrderStatus
	CreatedAt time.Time
	Status    EventStatus
}

type OrderCreatedEvent struct {
	UUID      string
	OrderUuid string
	CreatedAt time.Time
}
