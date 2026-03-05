package domain

import (
	"time"

	"github.com/nullableocean/grpcservices/shared/order"
)

type OrderUpdate struct {
	UUID      string
	Seq       int
	OrderUuid string
	NewStatus order.OrderStatus
	CreatedAt time.Time
}
