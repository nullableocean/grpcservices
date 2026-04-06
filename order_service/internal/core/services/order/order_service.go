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

var _ Service = &OrderService{}

type OrderService struct {
	orderRepo      ports.OrderRepository
	idemCache      ports.IdempotencyCache
	spotInstrument ports.SpotInstrument
	accessService  ports.AccessService
	metrics        ports.ServiceMetricsRecorder

	logger *zap.Logger
}

func NewOrderService(
	l *zap.Logger,
	oRepo ports.OrderRepository,
	spotInst ports.SpotInstrument,
	accessService ports.AccessService,
	metrics ports.ServiceMetricsRecorder,
	idemCache ports.IdempotencyCache,
) *OrderService {

	return &OrderService{
		idemCache:      idemCache,
		orderRepo:      oRepo,
		spotInstrument: spotInst,
		accessService:  accessService,
		metrics:        metrics,

		logger: l,
	}
}
