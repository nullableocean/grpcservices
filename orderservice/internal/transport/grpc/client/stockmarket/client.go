package stockmarket

import (
	"context"

	stockmarketv1 "github.com/nullableocean/grpcservices/api/gen/stockmarket/v1"
	"github.com/nullableocean/grpcservices/orderservice/internal/domain"
	"github.com/nullableocean/grpcservices/orderservice/internal/transport/mapping"
	"go.uber.org/zap"
)

type StockmarketClient struct {
	client stockmarketv1.StockMarketServiceClient
	logger *zap.Logger
}

func NewStockmarketClient(logger *zap.Logger, client stockmarketv1.StockMarketServiceClient) *StockmarketClient {
	return &StockmarketClient{
		client: client,
		logger: logger,
	}
}

func (c *StockmarketClient) ProcessOrder(ctx context.Context, o *domain.Order) error {
	req := mapping.MapDomainOrderToStockmarketProcessRequest(o)

	c.logger.Info("send order for processing in stockmarket service")

	_, err := c.client.ProcessOrder(ctx, req)
	if err != nil {
		c.logger.Warn("error send order to stockmarket", zap.Error(err))
		return err
	}

	return nil
}
