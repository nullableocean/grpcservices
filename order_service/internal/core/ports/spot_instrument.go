package ports

import (
	"context"

	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
)

type SpotInstrument interface {
	ViewMarkets(ctx context.Context, userRoles []model.UserRole) ([]*model.Market, error)
	FindMarket(ctx context.Context, marketUuid string, userRoles []model.UserRole) (*model.Market, error)
}
