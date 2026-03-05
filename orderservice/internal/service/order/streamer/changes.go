package streamer

import "github.com/nullableocean/grpcservices/shared/order"

type Changes interface {
	ChangedOrder() string
	GetChange() any
	IsFinal() bool
}

type StatusChanges struct {
	OrderUuid     string
	NewStatus     order.OrderStatus
	IsFinalStatus bool
}

func (ch *StatusChanges) ChangedOrder() string {
	return ch.OrderUuid
}

func (ch *StatusChanges) GetChange() any {
	return ch.NewStatus
}

func (ch *StatusChanges) IsFinal() bool {
	return ch.IsFinalStatus
}
