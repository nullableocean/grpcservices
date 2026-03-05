package ram

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/nullableocean/grpcservices/orderservice/internal/domain"
	"github.com/nullableocean/grpcservices/orderservice/internal/errs"
)

type OrderStore struct {
	store  map[string]*domain.Order
	nextId atomic.Int64

	mu sync.RWMutex
}

func NewOrderStore() *OrderStore {
	return &OrderStore{
		store:  make(map[string]*domain.Order, 256),
		nextId: atomic.Int64{},
		mu:     sync.RWMutex{},
	}
}

func (s *OrderStore) GetAll(ctx context.Context) []*domain.Order {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]*domain.Order, 0, len(s.store))
	for _, o := range s.store {
		out = append(out, o)
	}

	return out
}

func (s *OrderStore) Get(ctx context.Context, id string) (*domain.Order, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	o, found := s.store[id]
	if !found {
		return nil, errs.ErrNotFound
	}

	return o, nil
}

func (s *OrderStore) Save(ctx context.Context, ord *domain.Order) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if ord.UUID == "" {
		return fmt.Errorf("empty uuid: %w", errs.ErrInvalidData)
	}

	s.store[ord.UUID] = ord

	return nil
}
