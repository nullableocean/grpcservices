package model

type OrderStatus string

const (
	OrderStatusCreated   OrderStatus = "created"
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusCompleted OrderStatus = "completed"
	OrderStatusCancelled OrderStatus = "canceled"
	OrderStatusRejected  OrderStatus = "rejected"
	OrderStatusUnknown   OrderStatus = "unknown"
)

func (s OrderStatus) IsValid() bool {
	switch s {
	case OrderStatusCreated, OrderStatusPending, OrderStatusCompleted, OrderStatusCancelled, OrderStatusRejected:
		return true
	}

	return false
}
