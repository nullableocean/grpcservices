package ports

import (
	"context"

	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
)

type SpotInstrument interface {
	FindMarket(ctx context.Context, marketUuid string) (*model.Market, error)
}
