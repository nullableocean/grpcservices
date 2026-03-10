package outside

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

type UpdateStatusEvent struct {
	UUID             string
	OrderUuid        string
	NewStatus        order.OrderStatus
	ProcessingStatus EventStatus
	UpdatedAt        time.Time
}
