package bus

import (
	"context"
	"sync"

	"github.com/nullableocean/grpcservices/orderservice/internal/service/events/inside"
)

type EventHandler interface {
	Handle(ctx context.Context, e inside.Event)
}

type EventBus struct {
	handlers map[string][]EventHandler
	mu       sync.RWMutex
}

func NewEventBus() *EventBus {
	return &EventBus{
		handlers: make(map[string][]EventHandler),
		mu:       sync.RWMutex{},
	}
}

func (b *EventBus) Dispatch(ctx context.Context, e inside.Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	handlers, ex := b.handlers[e.EventType()]
	if !ex {
		return
	}

	for _, h := range handlers {
		h.Handle(ctx, e)
	}
}

func (b *EventBus) RegisterHandler(ctx context.Context, eventType string, h EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[eventType] = append(b.handlers[eventType], h)
}
