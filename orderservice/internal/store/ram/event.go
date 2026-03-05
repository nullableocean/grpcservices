package ram

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/nullableocean/grpcservices/orderservice/internal/domain"
	"github.com/nullableocean/grpcservices/orderservice/internal/errs"
)

type UpdateEventStore interface {
	Save(ctx context.Context, event *domain.UpdateEvent) error
	Find(ctx context.Context, uuid string) (*domain.UpdateEvent, error)
}
type EventStore struct {
	store  map[string]*domain.UpdateEvent
	nextId atomic.Int64

	mu sync.RWMutex
}

func NewEventStore() *EventStore {
	return &EventStore{
		store:  make(map[string]*domain.UpdateEvent, 256),
		nextId: atomic.Int64{},
		mu:     sync.RWMutex{},
	}
}

func (s *EventStore) Save(ctx context.Context, event *domain.UpdateEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if event.UUID == "" {
		return fmt.Errorf("empty uuid: %w", errs.ErrInvalidData)
	}

	if _, ex := s.store[event.UUID]; ex {
		return errs.ErrAlreadyExist
	}

	s.store[event.UUID] = event

	return nil
}

func (s *EventStore) Update(ctx context.Context, event *domain.UpdateEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if event.UUID == "" {
		return fmt.Errorf("empty uuid: %w", errs.ErrInvalidData)
	}

	s.store[event.UUID] = event

	return nil
}

func (s *EventStore) Find(ctx context.Context, uuid string) (*domain.UpdateEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	event, ex := s.store[uuid]
	if !ex {
		return nil, errs.ErrNotFound
	}

	return event, nil
}
