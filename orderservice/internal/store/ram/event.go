package ram

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/nullableocean/grpcservices/orderservice/internal/errs"
	"github.com/nullableocean/grpcservices/orderservice/internal/service/events/outside"
)

type UpdateEventStore interface {
	Save(ctx context.Context, event *outside.UpdateStatusEvent) error
	Find(ctx context.Context, uuid string) (*outside.UpdateStatusEvent, error)
}
type EventStore struct {
	store  map[string]*outside.UpdateStatusEvent
	nextId atomic.Int64

	mu sync.RWMutex
}

func NewEventStore() *EventStore {
	return &EventStore{
		store:  make(map[string]*outside.UpdateStatusEvent, 256),
		nextId: atomic.Int64{},
		mu:     sync.RWMutex{},
	}
}

func (s *EventStore) Save(ctx context.Context, event *outside.UpdateStatusEvent) error {
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

func (s *EventStore) Update(ctx context.Context, event *outside.UpdateStatusEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if event.UUID == "" {
		return fmt.Errorf("empty uuid: %w", errs.ErrInvalidData)
	}

	s.store[event.UUID] = event

	return nil
}

func (s *EventStore) Find(ctx context.Context, uuid string) (*outside.UpdateStatusEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	event, ex := s.store[uuid]
	if !ex {
		return nil, errs.ErrNotFound
	}

	return event, nil
}
