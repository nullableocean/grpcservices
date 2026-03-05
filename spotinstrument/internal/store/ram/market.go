package ram

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/nullableocean/grpcservices/spotinstrumentinstrument/internal/domain"
	"github.com/nullableocean/grpcservices/spotinstrumentinstrument/internal/service"
)

type MStore struct {
	markets map[string]*domain.Market
	nextId  atomic.Int64

	mu sync.RWMutex
}

func NewMarketStore() *MStore {
	return &MStore{
		markets: make(map[string]*domain.Market, 128),
		nextId:  atomic.Int64{},
		mu:      sync.RWMutex{},
	}
}

func (s *MStore) Save(ctx context.Context, market *domain.Market) (*domain.Market, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if market.UUID == "" {
		return nil, service.ErrInvalidUUID
	}

	s.markets[market.UUID] = market
	return market, nil
}

func (s *MStore) Get(ctx context.Context, uuid string) (*domain.Market, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	m, found := s.markets[uuid]
	if !found {
		return nil, fmt.Errorf("%w:market not found", service.ErrNotFound)
	}

	return m, nil
}

func (s *MStore) GetAll(ctx context.Context) []*domain.Market {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]*domain.Market, 0, len(s.markets))

	for _, v := range s.markets {
		out = append(out, v)
	}

	return out
}

func (s *MStore) GetEnabled(ctx context.Context) []*domain.Market {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]*domain.Market, 0, len(s.markets))

	for _, v := range s.markets {
		if v.IsEnabled() {
			out = append(out, v)
		}
	}

	return out
}

func (s *MStore) Delete(ctx context.Context, uuid string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	m, found := s.markets[uuid]
	if !found {
		return fmt.Errorf("%w:market not found", service.ErrNotFound)
	}

	delete(s.markets, uuid)
	m.Delete()

	return nil
}
