package stockmarket

import (
	"context"

	"github.com/nullableocean/grpcservices/orderservice/internal/domain"
	transport "github.com/nullableocean/grpcservices/orderservice/internal/transport/grpc/client/stockmarket"
	"go.uber.org/zap"
)

type StockmarketService struct {
	client *transport.StockmarketClient
	logger *zap.Logger
}

func NewStockMarketService(logger *zap.Logger, stockMarketClient *transport.StockmarketClient) *StockmarketService {
	return &StockmarketService{
		client: stockMarketClient,
		logger: logger,
	}
}

func (sm *StockmarketService) Process(ctx context.Context, o *domain.Order) error {
	sm.logger.Info("send order on stock market", zap.String("order_id", o.Id()))

	err := sm.client.ProcessOrder(ctx, o)
	if err != nil {
		sm.logger.Warn("error send order on stock market", zap.String("order_id", o.Id()), zap.Error(err))

		return err
	}

	return nil
}
