package model

import (
	"encoding/json"
	"time"
)

type EventType string

func (t EventType) String() string {
	return string(t)
}

const (
	EVENT_ORDER_CREATED EventType = "order.created"
	EVENT_ORDER_UPDATED EventType = "order.updated"
)

type Event interface {
	ID() string
	OrderID() string
	EventType() EventType
	Payload() ([]byte, error)
}

type EventCreatedData struct {
	Order *Order `json:"order"`
}

type EventOrderCreated struct {
	UUID      string
	OrderUUID string
	Data      *EventCreatedData
}

func (e *EventOrderCreated) ID() string {
	return e.UUID
}

func (e *EventOrderCreated) OrderID() string {
	return e.OrderUUID
}

func (e *EventOrderCreated) EventType() EventType {
	return EVENT_ORDER_CREATED
}

func (e *EventOrderCreated) Payload() ([]byte, error) {
	return json.Marshal(e.Data)
}

type EventUpdatedData struct {
	NewStatus *OrderStatus `json:"new_status"`
	OldStatus *OrderStatus `json:"old_status"`
	UpdatedAt time.Time    `json:"updated_at"`
}

type EventOrderUpdated struct {
	UUID      string
	OrderUUID string
	Data      *EventUpdatedData
}

func (e *EventOrderUpdated) ID() string {
	return e.UUID
}

func (e *EventOrderUpdated) OrderID() string {
	return e.OrderUUID
}

func (e *EventOrderUpdated) EventType() EventType {
	return EVENT_ORDER_UPDATED
}

func (e *EventOrderUpdated) Payload() ([]byte, error) {
	return json.Marshal(e.Data)
}
