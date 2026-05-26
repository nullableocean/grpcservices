package order

import (
	"context"

	"github.com/nullableocean/grpcservices/orderservice/internal/core/dto"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/ports"
	"go.uber.org/zap"
)

type Service interface {
	CreateOrder(ctx context.Context, data *dto.CreateOrderParameters) (*model.Order, error)
	GetOrder(ctx context.Context, orderUUID, userUUID string) (*model.Order, error)
	UpdateOrder(ctx context.Context, orderUUID string, data *dto.UpdateOrderParameters) error
}

type OrderService struct {
	orderRepo     ports.OrderRepository
	accessService ports.AccessService
	metrics       ports.ServiceMetricsRecorder
	logger        *zap.Logger

	idempotencyGuard *IdempotencyGuard
	marketValidator  *MarketValidator
	orderFactory     *OrderFactory
}

func NewOrderService(
	logger *zap.Logger,
	orderRepo ports.OrderRepository,
	spotInstrument ports.SpotInstrument,
	accessService ports.AccessService,
	metrics ports.ServiceMetricsRecorder,
	idempotencyCache ports.IdempotencyCache,
) *OrderService {
	return &OrderService{
		orderRepo:        orderRepo,
		accessService:    accessService,
		metrics:          metrics,
		logger:           logger,
		idempotencyGuard: NewIdempotencyGuard(idempotencyCache, logger),
		marketValidator:  NewMarketValidator(spotInstrument, logger),
		orderFactory:     NewOrderFactory(),
	}
}
