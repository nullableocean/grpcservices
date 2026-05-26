package order

import (
	"time"

	"github.com/google/uuid"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/dto"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
)

type OrderFactory struct{}

func NewOrderFactory() *OrderFactory {
	return &OrderFactory{}
}

func (f *OrderFactory) CreateOrder(data *dto.CreateOrderParameters) *model.Order {
	now := time.Now()
	return &model.Order{
		UUID:       uuid.NewString(),
		UserUUID:   data.User.UUID,
		MarketUUID: data.MarketUUID,
		Status:     model.OrderStatusCreated,
		Type:       data.Type,
		Side:       data.Side,
		Price:      data.Price,
		Quantity:   data.Quantity,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

func (f *OrderFactory) CreateCreatedEvent(order *model.Order) *model.EventOrderCreated {
	return &model.EventOrderCreated{
		UUID:      uuid.NewString(),
		OrderUUID: order.UUID,
		Data: &model.EventCreatedData{
			Order: order,
		},
	}
}

func (f *OrderFactory) CreateUpdatedEvent(orderUUID string, oldStatus, newStatus model.OrderStatus) *model.EventOrderUpdated {
	return &model.EventOrderUpdated{
		UUID:      uuid.NewString(),
		OrderUUID: orderUUID,
		Data: &model.EventUpdatedData{
			NewStatus: &newStatus,
			OldStatus: &oldStatus,
			UpdatedAt: time.Now(),
		},
	}
}
