package model

import (
	"encoding/json"
	"fmt"
)

type OrderStatus string

const (
	OrderStatusCreated   OrderStatus = "CREATED"
	OrderStatusPending   OrderStatus = "PENDING"
	OrderStatusCompleted OrderStatus = "COMPLETED"
	OrderStatusCancelled OrderStatus = "CANCELLED"
	OrderStatusRejected  OrderStatus = "REJECTED"
)

func (s OrderStatus) IsValid() bool {
	switch s {
	case OrderStatusCreated, OrderStatusPending, OrderStatusCompleted, OrderStatusCancelled, OrderStatusRejected:
		return true
	}

	return false
}

func AllowedTransitions(current OrderStatus) []OrderStatus {
	switch current {
	case OrderStatusCreated:
		return []OrderStatus{OrderStatusPending, OrderStatusRejected, OrderStatusCancelled}
	case OrderStatusPending:
		return []OrderStatus{OrderStatusCompleted, OrderStatusCancelled, OrderStatusRejected}
	default:
		return []OrderStatus{}
	}
}

func (s OrderStatus) CanTransitTo(new OrderStatus) bool {
	for _, allowedStatus := range AllowedTransitions(s) {
		if allowedStatus == new {
			return true
		}
	}

	return false
}

func (s OrderStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(s))
}

func (s *OrderStatus) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	status := OrderStatus(str)
	if !status.IsValid() {
		return fmt.Errorf("invalid order status: %s", str)
	}

	*s = status
	return nil
}
