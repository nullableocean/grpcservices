package ram

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/nullableocean/grpcservices/order/domain"
)

type EventStore struct {
	store  map[int64]*domain.UpdateEvent
	nextId atomic.Int64

	mu sync.RWMutex
}

func NewEventStore() *EventStore {
	return &EventStore{
		store:  make(map[int64]*domain.UpdateEvent, 256),
		nextId: atomic.Int64{},
		mu:     sync.RWMutex{},
	}
}

func (s *EventStore) Save(ctx context.Context, event *domain.UpdateEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := s.nextId.Add(1)
	event.Id = id

	s.store[id] = event

	return nil
}
