package service

import (
	"errors"
	"main/pkg/roles"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrNotFoundMarket = errors.New("not found market")
)

type SpotInstrument struct {
	markets []*Market
	nextId  atomic.Int64

	mu sync.RWMutex
}

func NewSpotInstrument() *SpotInstrument {
	return &SpotInstrument{
		markets: []*Market{},
		mu:      sync.RWMutex{},
	}
}

func (s *SpotInstrument) ViewMarkets(roles []roles.UserRole) []*Market {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]*Market, 0, len(s.markets))

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

func (s *SpotInstrument) NewMarket(name string, allowed []roles.UserRole) *Market {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := s.nextId.Add(1)

	allowedMap := make(map[roles.UserRole]struct{}, len(allowed))
	for _, r := range allowed {
		allowedMap[r] = struct{}{}
	}

	newMarket := &Market{
		id:           id,
		name:         name,
		enabled:      true,
		deletedAt:    nil,
		allowedRoles: allowedMap,
	}

	s.markets = append(s.markets, newMarket)

	return newMarket
}

func (s *SpotInstrument) DeleteMarket(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	found := false
	for _, m := range s.markets {
		now := time.Now()
		if m.Id() == id {
			m.deletedAt = &now
			m.enabled = false

			found = true
			break
		}
	}

	if !found {
		return ErrNotFoundMarket
	}

	return nil
}
