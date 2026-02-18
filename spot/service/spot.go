package service

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	"github.com/nullableocean/grpcservices/pkg/roles"
	"github.com/nullableocean/grpcservices/spot/domain"
)

var (
	ErrNotFoundMarket = errors.New("not found market")
)

type SpotInstrument struct {
	markets []*domain.Market
	nextId  atomic.Int64

	mu sync.RWMutex
}

func NewSpotInstrument() *SpotInstrument {
	return &SpotInstrument{
		markets: []*domain.Market{},
		mu:      sync.RWMutex{},
	}
}

func (s *SpotInstrument) ViewMarkets(ctx context.Context, roles []roles.UserRole) []*domain.Market {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]*domain.Market, 0, len(s.markets))

	for _, m := range s.markets {
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

func (s *SpotInstrument) NewMarket(name string, allowed []roles.UserRole) (*domain.Market, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := s.nextId.Add(1)

	allowedMap := make(map[roles.UserRole]struct{}, len(allowed))
	for _, r := range allowed {
		allowedMap[r] = struct{}{}
	}

	newMarket := domain.NewMarket(domain.CreateMarketDto{
		Id:           id,
		Name:         name,
		Enabled:      true,
		AllowedRoles: allowedMap,
	})

	s.markets = append(s.markets, newMarket)

	return newMarket, nil
}

func (s *SpotInstrument) DeleteMarket(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	found := false
	for _, m := range s.markets {
		if m.Id() == id {
			m.Delete()

			found = true
			break
		}
	}

	if !found {
		return ErrNotFoundMarket
	}

	return nil
}
