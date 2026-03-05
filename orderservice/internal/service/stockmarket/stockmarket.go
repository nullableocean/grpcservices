package stockmarket

import (
	"context"

	"github.com/nullableocean/grpcservices/orderservice/internal/domain"
	"github.com/nullableocean/grpcservices/orderservice/internal/service/order"
	transport "github.com/nullableocean/grpcservices/orderservice/internal/transport/grpc/client/stockmarket"
	"go.uber.org/zap"
)

var _ order.Processor = &GrpcStockmarketService{}

type GrpcStockmarketService struct {
	client *transport.StockmarketClient
	logger *zap.Logger
}

func NewStockMarketService(logger *zap.Logger, stockMarketClient *transport.StockmarketClient) *GrpcStockmarketService {
	return &GrpcStockmarketService{
		client: stockMarketClient,
		logger: logger,
	}
}

func (sm *GrpcStockmarketService) Process(ctx context.Context, o *domain.Order) error {
	sm.logger.Info("send order on stock market", zap.String("order_id", o.Id()))

	err := sm.client.ProcessOrder(ctx, o)
	if err != nil {
		sm.logger.Warn("error send order on stock market", zap.String("order_id", o.Id()), zap.Error(err))

		return err
	}

	return nil
}
