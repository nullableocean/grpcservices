package ports

import (
	"context"

	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
)

type Sub interface {
	Updates() <-chan *model.EventOrderUpdated
	Close()
}

type UpdateNotifier interface {
	Subscribe(ctx context.Context, orderUUID string) Sub
}
