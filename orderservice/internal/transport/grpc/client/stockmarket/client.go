package stockmarket

import (
	"context"

	stockmarketv1 "github.com/nullableocean/grpcservices/api/gen/stockmarket/v1"
	"github.com/nullableocean/grpcservices/orderservice/internal/domain"
	"github.com/nullableocean/grpcservices/orderservice/internal/transport/mapping"
	"go.opentelemetry.io/otel"
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
	ctx, span := otel.Tracer("stockmarket_client").Start(ctx, "process_order_request")
	defer span.End()

	req := mapping.MapDomainOrderToStockmarketProcessRequest(o)

	c.logger.Info("send request in stockmarket grpc server")

	_, err := c.client.ProcessOrder(ctx, req)
	if err != nil {
		c.logger.Error("failed send order to stockmarket", zap.Error(err))
		return err
	}

	return nil
}
