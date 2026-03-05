package order

import (
	"context"

	"github.com/nullableocean/grpcservices/orderservice/internal/domain"
)

type Processor interface {
	Process(ctx context.Context, o *domain.Order) error
}
