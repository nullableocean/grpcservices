package market

import (
	"context"
	"math/rand"
	"time"

	"github.com/nullableocean/grpcservices/stockmarketservice/internal/domain"
	"github.com/nullableocean/grpcservices/stockmarketservice/internal/service/errs"
	"go.opentelemetry.io/otel"
)

type MarketService struct{}

func NewMarketService() *MarketService {
	return &MarketService{}
}

func (s *MarketService) Buy(ctx context.Context, o *domain.Order) error {
	_, span := otel.Tracer("stockmarket_market_service").Start(ctx, "order_buy_process")
	defer span.End()

	// dummy
	waiting := rand.Intn(5)

	select {
	case <-time.After(time.Duration(waiting) * time.Second):
	case <-ctx.Done():
		return ctx.Err()
	}

	chance := rand.Intn(100)
	if chance < 28 {
		return errs.ErrDontWantProcessOrder
	}

	return nil
}

func (s *MarketService) Sell(ctx context.Context, o *domain.Order) error {
	_, span := otel.Tracer("stockmarket_market_service").Start(ctx, "order_sell_process")
	defer span.End()

	// dummy
	// dummy
	waiting := rand.Intn(5)

	select {
	case <-time.After(time.Duration(waiting) * time.Second):
	case <-ctx.Done():
		return ctx.Err()
	}

	chance := rand.Intn(100)
	if chance < 28 {
		return errs.ErrDontWantProcessOrder
	}

	return nil
}
