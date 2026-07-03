package spotinstrument

import (
	"context"
	"fmt"

	"github.com/nullableocean/grpcservices/shared/logger"
	"github.com/nullableocean/grpcservices/spotinstrument/internal/core/errs"
	"github.com/nullableocean/grpcservices/spotinstrument/internal/core/model"
	"github.com/nullableocean/grpcservices/spotinstrument/internal/core/ports/metrics"
	"github.com/nullableocean/grpcservices/spotinstrument/internal/core/ports/repository"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

type SpotInstrument struct {
	marketRepo repository.MarketRepository
	metrics    metrics.SpotInstrumentRecords

	logger *logger.CtxZapLogger
}

func NewSpotInstrument(l *logger.CtxZapLogger, mRepo repository.MarketRepository, metrics metrics.SpotInstrumentRecords) *SpotInstrument {
	return &SpotInstrument{
		marketRepo: mRepo,
		metrics:    metrics,
		logger:     l,
	}
}

func (s *SpotInstrument) ViewMarketsPaginated(ctx context.Context, user *model.User, pageToken model.PageToken, pageSize int32) (*model.PaginationData, error) {
	ctx, span := otel.Tracer("spot_instrument").Start(ctx, "view_markets")
	defer span.End()

	s.metrics.ViewMarkets(ctx)
	s.logger.Debug(ctx, "view markets with pagination", zap.String("page_token", pageToken.Token))

	paginatonData, err := s.marketRepo.FindEnabledByRolesPaginated(ctx, user.Roles, pageToken, pageSize)
	if err != nil {
		span.AddEvent("failed get markets")
		s.metrics.FailedViewMarkets(ctx)

		return nil, fmt.Errorf("failed to get markets: %w", err)
	}

	return paginatonData, nil
}

func (s *SpotInstrument) ViewMarkets(ctx context.Context, userRoles []model.UserRole) ([]*model.Market, error) {
	ctx, span := otel.Tracer("spot_instrument").Start(ctx, "view_markets")
	defer span.End()

	s.metrics.ViewMarkets(ctx)
	s.logger.Debug(ctx, "view markets")

	markets, err := s.marketRepo.FindEnabledByRoles(ctx, userRoles)
	if err != nil {
		span.AddEvent("failed get markets")
		s.metrics.FailedViewMarkets(ctx)

		return nil, fmt.Errorf("failed to get markets: %w", err)
	}

	return markets, nil
}

func (s *SpotInstrument) FindWithRoles(ctx context.Context, marketUuid string, userRoles []model.UserRole) (*model.Market, error) {
	ctx, span := otel.Tracer("spot_instrument").Start(ctx, "find_market_with_roles")
	defer span.End()
	span.SetAttributes(attribute.String("market_uuid", marketUuid))

	s.metrics.FindMarket(ctx)
	s.logger.Debug(ctx, "find market", zap.String("market_uuid", marketUuid))

	market, err := s.marketRepo.FindByUUID(ctx, marketUuid)
	if err != nil {
		s.metrics.FailedFindMarket(ctx)
		s.logger.Error(ctx, "failed find market", zap.Error(err), zap.String("market_uuid", marketUuid))

		return nil, fmt.Errorf("failed to get market: %w", err)
	}

	if !market.IsAccessibleForRoles(userRoles) {
		span.AddEvent("failed find market")
		s.logger.Error(ctx, "market found. not allowed for roles", zap.String("market_uuid", marketUuid))

		return nil, errs.ErrNotAllowed
	}

	return market, nil
}
