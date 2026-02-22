package ram

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/nullableocean/grpcservices/order/domain"
	"github.com/nullableocean/grpcservices/order/service"
	"github.com/nullableocean/grpcservices/pkg/order"
)

type OrderStore struct {
	store  map[int64]*domain.Order
	nextId atomic.Int64

	mu sync.RWMutex
}

func NewOrderStore() *OrderStore {
	return &OrderStore{
		store:  make(map[int64]*domain.Order, 256),
		nextId: atomic.Int64{},
		mu:     sync.RWMutex{},
	}
}

func (s *OrderStore) Get(ctx context.Context, id int64) (*domain.Order, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	o, found := s.store[id]
	if !found {
		return nil, service.ErrNotFound
	}

	return o, nil
}

func (s *OrderStore) Create(ctx context.Context, orderData *domain.CreateOrderDto) (*domain.Order, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := s.nextId.Add(1)
	newOrder := domain.NewOrder(id, orderData)

	s.store[id] = newOrder

	return newOrder, nil
}

func (s *OrderStore) UpdateStatus(ctx context.Context, order *domain.Order, newStatus order.OrderStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	o, found := s.store[order.Id()]
	if !found {
		return service.ErrNotFound
	}
	o.SetStatus(newStatus)

	return nil
}
