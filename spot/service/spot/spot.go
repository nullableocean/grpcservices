package spot

import (
	"context"

	"github.com/nullableocean/grpcservices/pkg/roles"
	"github.com/nullableocean/grpcservices/spot/domain"
)

type MarketStore interface {
	Save(ctx context.Context, marketData *domain.CreateMarketDto) (*domain.Market, error)
	GetById(ctx context.Context, id int64) (*domain.Market, error)
	GetAll(ctx context.Context) []*domain.Market
	DeleteById(ctx context.Context, id int64) error
}

type SpotInstrument struct {
	store MarketStore
}

func NewSpotInstrument(store MarketStore) *SpotInstrument {
	return &SpotInstrument{
		store: store,
	}
}

func (s *SpotInstrument) ViewMarkets(ctx context.Context, roles []roles.UserRole) []*domain.Market {
	markets := s.store.GetAll(ctx)
	out := make([]*domain.Market, 0, len(markets))

	for _, m := range markets {
	ROLE_LOOP:
		for _, r := range roles {
			if m.IsAllowed(r) && m.IsEnabled() && !m.IsDeleted() {
				out = append(out, m)
				break ROLE_LOOP
			}
		}
	}

	return out
}

func (s *SpotInstrument) NewMarket(ctx context.Context, name string, allowed []roles.UserRole) (*domain.Market, error) {
	allowedMap := make(map[roles.UserRole]struct{}, len(allowed))
	for _, r := range allowed {
		allowedMap[r] = struct{}{}
	}

	dto := &domain.CreateMarketDto{
		Name:         name,
		Enabled:      true,
		AllowedRoles: allowedMap,
	}

	newMarket, err := s.store.Save(ctx, dto)
	return newMarket, err
}

func (s *SpotInstrument) DeleteMarket(ctx context.Context, id int64) error {
	return s.store.DeleteById(ctx, id)
}
