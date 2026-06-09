package updatenotifier

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/ports"
	"go.uber.org/zap"
)

var _ ports.UpdateNotifier = &UpdateNotifier{}
var _ ports.EventPublisher = &UpdateNotifier{}

type UpdateNotifier struct {
	subs   map[string]*Subs
	nextId int

	sendTimeout time.Duration
	sendTries   int32

	mu sync.RWMutex

	logger *zap.Logger
}

type Options struct {
	SendTimeoutOnSub time.Duration
	SendTries        int
}

func NewUpdateNotifier(l *zap.Logger, opt Options) (*UpdateNotifier, error) {
	if opt.SendTimeoutOnSub <= 0 {
		return nil, errors.New("invalid timeout option")
	}

	if opt.SendTries <= 0 {
		return nil, errors.New("invalid tries option")
	}

	return &UpdateNotifier{
		subs:        map[string]*Subs{},
		sendTimeout: opt.SendTimeoutOnSub,
		sendTries:   int32(opt.SendTries),
		mu:          sync.RWMutex{},
		logger:      l,
	}, nil
}

func (notifier *UpdateNotifier) Subscribe(ctx context.Context, orderUUID string) ports.Sub {
	notifier.mu.Lock()
	defer notifier.mu.Unlock()

	newSub := &Sub{
		updatesCh: make(chan *model.EventOrderUpdated, 1),
		closeCh:   make(chan struct{}),
		closeOnce: sync.Once{},
	}

	if _, ex := notifier.subs[orderUUID]; !ex {
		notifier.subs[orderUUID] = NewSubs()
	}

	notifier.subs[orderUUID].Add(newSub)

	return newSub
}

func (notifier *UpdateNotifier) Publish(ctx context.Context, event model.Event) error {
	updatedEvent, ok := event.(*model.EventOrderUpdated)
	if !ok {
		notifier.logger.Warn("update notifier got not updated event", zap.String("event_type", event.EventType().String()))

		return nil
	}

	err := notifier.publish(ctx, event.GetOrderUUID(), updatedEvent)
	if err != nil {
		return err
	}

	return nil
}

func (notifier *UpdateNotifier) publish(ctx context.Context, orderUUID string, event *model.EventOrderUpdated) error {
	logger := notifier.logger.With(zap.String("event_uuid", event.UUID), zap.String("order_uuid", event.OrderUUID))

	notifier.mu.RLock()
	subs, ex := notifier.subs[orderUUID]
	if !ex {
		notifier.mu.RUnlock()

		return nil
	}
	notifier.mu.RUnlock()

	cpSubs := subs.GetSubs()
	for _, sub := range cpSubs {
		select {
		case <-ctx.Done():
			logger.Info("context closed", zap.Error(ctx.Err()))
			return ctx.Err()

		case <-sub.closeCh:
			logger.Info("sub closed", zap.Int("sub_id", sub.id))

			subs.Remove(sub.id)
		case <-time.After(notifier.sendTimeout):
			logger.Info("timeout notify sub", zap.Int("sub_id", sub.id))

			if atomic.AddInt32(&sub.timeouts, 1) >= notifier.sendTries {
				logger.Info("expired timeouts. remove sub", zap.Int("sub_id", sub.id))

				subs.Remove(sub.id)
			}
		case sub.updatesCh <- event:
			logger.Info("subscriber notified", zap.Int("sub_id", sub.id))
		}
	}

	return nil
}
