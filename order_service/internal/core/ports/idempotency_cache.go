package ports

import (
	"context"

	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
)

type IdempotencyCache interface {
	Get(ctx context.Context, key string) (*model.IdempotencyData, error)
	SetIfNotExist(ctx context.Context, key string, data *model.IdempotencyData) (bool, error)
	Update(ctx context.Context, key string, data *model.IdempotencyData) error
	Delete(ctx context.Context, key string) error
}
