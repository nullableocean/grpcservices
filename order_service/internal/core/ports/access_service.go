package ports

import (
	"context"

	"github.com/nullableocean/grpcservices/orderservice/internal/core/dto"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
)

type AccessService interface {
	CanCreateOrder(ctx context.Context, user *model.User, createParams *dto.CreateOrderParameters) error
}
