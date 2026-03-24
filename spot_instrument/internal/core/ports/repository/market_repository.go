package repository

import (
	"context"

	"github.com/nullableocean/grpcservices/spotinstrument/internal/core/model"
)

type MarketRepository interface {
	FindEnabledByRoles(ctx context.Context, roles []model.UserRole) ([]*model.Market, error)
	FindByUUID(ctx context.Context, uuid string) (*model.Market, error)
	Create(ctx context.Context, market *model.Market) error
	Update(ctx context.Context, market *model.Market) error
	Delete(ctx context.Context, uuid string) error
}
