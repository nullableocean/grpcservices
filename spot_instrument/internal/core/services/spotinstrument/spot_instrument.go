package spotinstrument

import (
	"context"
	"fmt"

	"github.com/nullableocean/grpcservices/spotinstrument/internal/core/model"
	"github.com/nullableocean/grpcservices/spotinstrument/internal/core/ports/metrics"
	"github.com/nullableocean/grpcservices/spotinstrument/internal/core/ports/repository"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

type SpotInstrument struct {
	marketRepo repository.MarketRepository
	metrics    metrics.SpotInstrumentRecords

	logger *zap.Logger
}

func NewSpotInstrument(l *zap.Logger, mRepo repository.MarketRepository, metrics metrics.SpotInstrumentRecords) *SpotInstrument {
	return &SpotInstrument{
		marketRepo: mRepo,
		metrics:    metrics,
		logger:     l,
	}
}

func (s *SpotInstrument) ViewMarkets(ctx context.Context, userRoles []model.UserRole) ([]*model.Market, error) {
	ctx, span := otel.Tracer("spot_instrument").Start(ctx, "view_markets")
	defer span.End()

	s.metrics.ViewMarkets(ctx)
	s.logger.Info("view markets")

	markets, err := s.marketRepo.FindEnabledByRoles(ctx, userRoles)
	if err != nil {
		s.metrics.FailedViewMarkets(ctx)
		return nil, fmt.Errorf("failed to get markets: %w", err)
	}

	return markets, nil
}
