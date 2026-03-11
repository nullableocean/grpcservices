package spot

import (
	"context"

	"github.com/google/uuid"
	"github.com/nullableocean/grpcservices/shared/roles"
	"github.com/nullableocean/grpcservices/spotinstrumentinstrument/internal/domain"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

type RoleAccess interface {
	HasAccessToMarket(m *domain.Market, role roles.UserRole) bool
}

type MarketStore interface {
	Save(ctx context.Context, newMarket *domain.Market) (*domain.Market, error)
	Get(ctx context.Context, uuid string) (*domain.Market, error)
	GetAll(ctx context.Context) ([]*domain.Market, error)
	Delete(ctx context.Context, uuid string) error
}

type SpotInstrument struct {
	store      MarketStore
	roleAccess RoleAccess
	logger     *zap.Logger
}

func NewSpotInstrument(logger *zap.Logger, store MarketStore, roleAccessService RoleAccess) *SpotInstrument {
	return &SpotInstrument{
		store:      store,
		roleAccess: roleAccessService,

		logger: logger,
	}
}

func (s *SpotInstrument) ViewMarkets(ctx context.Context, roles []roles.UserRole) ([]*domain.Market, error) {
	ctx, span := otel.Tracer("spotinstrument_service").Start(ctx, "view_markets")
	defer span.End()

	s.logger.Info("get markets from store")

	markets, err := s.store.GetAll(ctx)
	if err != nil {
		span.AddEvent("failed get markets")
		s.logger.Error("failed get markets from store")

		return nil, err
	}

	out := make([]*domain.Market, 0, len(markets))

	s.logger.Info("get allowed markets for user roles")
	for _, m := range markets {
	ROLE_LOOP:
		for _, r := range roles {
			if m.IsEnabled() && s.roleAccess.HasAccessToMarket(m, r) {
				out = append(out, m)
				break ROLE_LOOP
			}
		}
	}

	return out, nil
}

func (s *SpotInstrument) NewMarket(ctx context.Context, dto *domain.CreateMarketDto) (*domain.Market, error) {
	ctx, span := otel.Tracer("spotinstrument_service").Start(ctx, "new_market")
	defer span.End()

	allowedMap := make(map[roles.UserRole]struct{}, len(dto.AllowedRoles))
	for _, r := range dto.AllowedRoles {
		allowedMap[r] = struct{}{}
	}

	newMarket := &domain.Market{
		UUID:         uuid.NewString(),
		Name:         dto.Name,
		Enabled:      dto.Enabled,
		AllowedRoles: roles.NewRoles(dto.AllowedRoles...),
		DeletedAt:    nil,
	}

	s.logger.Info("store new market")
	newMarket, err := s.store.Save(ctx, newMarket)
	if err != nil {
		span.AddEvent("failed store market")
		s.logger.Error("failed store new market", zap.Error(err))

		return nil, err
	}

	return newMarket, nil
}

func (s *SpotInstrument) DeleteMarket(ctx context.Context, uuid string) error {
	ctx, span := otel.Tracer("spotinstrument_service").Start(ctx, "delete_market")
	defer span.End()
	s.logger.Info("delete market", zap.String("market_uuid", uuid))

	return s.store.Delete(ctx, uuid)
}
