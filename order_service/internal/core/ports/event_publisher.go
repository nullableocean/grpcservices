package ports

import (
	"context"

	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
)

type EventPublisher interface {
	Publish(ctx context.Context, event model.Event) error
}

type CreatedEventPublisher interface {
	Publish(ctx context.Context, event *model.EventOrderCreated) error
}

type UpdatedEventPublisher interface {
	Publish(ctx context.Context, event *model.EventOrderUpdated) error
}
