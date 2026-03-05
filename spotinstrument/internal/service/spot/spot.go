package spot

import (
	"context"

	"github.com/google/uuid"
	"github.com/nullableocean/grpcservices/shared/roles"
	"github.com/nullableocean/grpcservices/spotinstrumentinstrument/internal/domain"
)

type RoleAccess interface {
	HasAccessToMarket(m *domain.Market, role roles.UserRole) bool
}

type MarketStore interface {
	Save(ctx context.Context, newMarket *domain.Market) (*domain.Market, error)
	Get(ctx context.Context, uuid string) (*domain.Market, error)
	GetAll(ctx context.Context) []*domain.Market
	Delete(ctx context.Context, uuid string) error
}

type SpotInstrument struct {
	store      MarketStore
	roleAccess RoleAccess
}

func NewSpotInstrument(store MarketStore, roleAccessService RoleAccess) *SpotInstrument {
	return &SpotInstrument{
		store:      store,
		roleAccess: roleAccessService,
	}
}

func (s *SpotInstrument) ViewMarkets(ctx context.Context, roles []roles.UserRole) []*domain.Market {
	markets := s.store.GetAll(ctx)
	out := make([]*domain.Market, 0, len(markets))

	for _, m := range markets {
	ROLE_LOOP:
		for _, r := range roles {
			if m.IsEnabled() && s.roleAccess.HasAccessToMarket(m, r) {
				out = append(out, m)
				break ROLE_LOOP
			}
		}
	}

	return out
}

func (s *SpotInstrument) NewMarket(ctx context.Context, dto *domain.CreateMarketDto) (*domain.Market, error) {
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

	newMarket, err := s.store.Save(ctx, newMarket)
	if err != nil {
		return nil, err
	}

	return newMarket, nil
}

func (s *SpotInstrument) DeleteMarket(ctx context.Context, uuid string) error {
	return s.store.Delete(ctx, uuid)
}
