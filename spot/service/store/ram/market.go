package ram

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/nullableocean/grpcservices/spot/domain"
	"github.com/nullableocean/grpcservices/spot/service"
)

type MStore struct {
	markets map[int64]*domain.Market
	nextId  atomic.Int64

	mu sync.RWMutex
}

func NewMarketStore() *MStore {
	return &MStore{
		markets: make(map[int64]*domain.Market, 128),
		nextId:  atomic.Int64{},
		mu:      sync.RWMutex{},
	}
}

func (s *MStore) Save(ctx context.Context, marketData *domain.CreateMarketDto) (*domain.Market, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := s.nextId.Add(1)

	newMarket := domain.NewMarket(id, marketData)
	s.markets[id] = newMarket

	return newMarket, nil
}

func (s *MStore) GetById(ctx context.Context, id int64) (*domain.Market, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	m, found := s.markets[id]
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

func (s *MStore) DeleteById(ctx context.Context, id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	m, found := s.markets[id]
	if !found {
		return fmt.Errorf("%w:market not found", service.ErrNotFound)
	}

	delete(s.markets, id)
	m.Delete()

	return nil
}
