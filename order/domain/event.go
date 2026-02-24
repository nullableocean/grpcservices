package domain

import (
	"time"

	"github.com/nullableocean/grpcservices/pkg/order"
)

type UpdateEvent struct {
	Id        int64
	OrderId   int64
	NewStatus order.OrderStatus
	CreatedAt time.Time
}

type MarketEvent struct {
	Id        int64
	OrderId   int64
	NewStatus order.OrderStatus
	CreatedAt time.Time
}
