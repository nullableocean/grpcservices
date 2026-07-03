package order

import (
	"context"

	"github.com/nullableocean/grpcservices/orderservice/internal/core/dto"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/ports"
	"github.com/nullableocean/grpcservices/shared/logger"
)

type Service interface {
	OrdersList(ctx context.Context, user *model.User, filters model.OrderListFilter) (model.OrderList, error)
	CreateOrder(ctx context.Context, data *dto.CreateOrderParameters) (*model.Order, error)
	GetOrder(ctx context.Context, orderUUID string, user *model.User) (*model.Order, error)
	UpdateOrder(ctx context.Context, orderUUID string, data *dto.UpdateOrderParameters) error
}

type OrderService struct {
	orderRepo     ports.OrderRepository
	accessService ports.AccessService
	metrics       ports.ServiceMetricsRecorder
	rateLimiter   ports.OrderRateLimiter
	logger        *logger.CtxZapLogger

	idempotencyGuard *IdempotencyGuard
	marketValidator  *MarketValidator
	orderFactory     *OrderFactory
}

func NewOrderService(
	logger *logger.CtxZapLogger,
	orderRepo ports.OrderRepository,
	spotInstrument ports.SpotInstrument,
	accessService ports.AccessService,
	metrics ports.ServiceMetricsRecorder,
	rateLimiter ports.OrderRateLimiter,
	idempotencyCache ports.IdempotencyCache,
) *OrderService {
	return &OrderService{
		orderRepo:        orderRepo,
		accessService:    accessService,
		metrics:          metrics,
		rateLimiter:      rateLimiter,
		logger:           logger,
		idempotencyGuard: NewIdempotencyGuard(idempotencyCache, logger),
		marketValidator:  NewMarketValidator(spotInstrument, logger),
		orderFactory:     NewOrderFactory(),
	}
}
