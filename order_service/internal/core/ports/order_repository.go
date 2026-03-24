package ports

import (
	"context"

	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
)

type OrderRepository interface {
	Save(ctx context.Context, order *model.Order, events ...model.Event) error
	Update(ctx context.Context, updatedOrder *model.Order, events ...model.Event) error
	FindByUUID(ctx context.Context, orderUUID string) (*model.Order, error)
}
