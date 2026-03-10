package bus

import (
	"context"
	"sync"

	"github.com/nullableocean/grpcservices/orderservice/internal/service/events/inside"
	"github.com/nullableocean/grpcservices/shared/limiter"
	"go.uber.org/zap"
)

var (
	defaultProcLimit = 5
)

type EventHandler interface {
	Handle(ctx context.Context, e inside.Event)
}

type EventBus struct {
	handlers map[string][]EventHandler
	limiter  *limiter.Limiter
	mu       sync.RWMutex

	logger *zap.Logger
}

type Option struct {
	HandleProcessLimit int
}

func NewEventBus(logger *zap.Logger, opt Option) *EventBus {
	if opt.HandleProcessLimit <= 0 {
		opt.HandleProcessLimit = defaultProcLimit
	}

	return &EventBus{
		handlers: make(map[string][]EventHandler),
		limiter:  limiter.New(opt.HandleProcessLimit),
		mu:       sync.RWMutex{},
		logger:   logger,
	}
}

func (b *EventBus) Dispatch(ctx context.Context, e inside.Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	handlers, ex := b.handlers[e.EventType()]
	if !ex {
		return
	}

	b.limiter.Acquire()
	ctx = context.WithoutCancel(ctx)

	go func() {
		defer b.limiter.Release()
		defer b.handlePanic()

		for _, h := range handlers {
			h.Handle(ctx, e)
		}
	}()
}

func (b *EventBus) RegisterHandler(ctx context.Context, eventType string, h EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[eventType] = append(b.handlers[eventType], h)
}

func (b *EventBus) handlePanic() {
	if r := recover(); r != nil {
		b.logger.Error("panic in event bus", zap.Any("recover", r), zap.Stack("stack"))
	}
}
