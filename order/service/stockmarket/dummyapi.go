package stockmarket

import (
	"context"

	"github.com/nullableocean/grpcservices/order/domain"
)

type DummyMarketClient struct{}

func NewDummyMarketClient() *DummyMarketClient {
	return &DummyMarketClient{}
}

func (c *DummyMarketClient) CreateMarketOrder(ctx context.Context, o *domain.Order) error {
	return nil
}
