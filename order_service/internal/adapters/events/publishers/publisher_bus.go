package publishers

import (
	"context"
	"errors"

	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/ports"
)

// var _ ports.EventPublisher = &EventPublisherBus{}

type EventPublisherBus struct {
	publishers map[model.EventType][]ports.EventPublisher
}

func NewEventPublisherBus() *EventPublisherBus {
	return &EventPublisherBus{
		publishers: map[model.EventType][]ports.EventPublisher{},
	}
}

func (bus *EventPublisherBus) Register(eventType model.EventType, publisher ports.EventPublisher) {
	bus.publishers[eventType] = append(bus.publishers[eventType], publisher)
}

func (p *EventPublisherBus) Publish(ctx context.Context, e model.Event) error {
	publishers := p.publishers[e.EventType()]
	if len(publishers) == 0 {
		return nil
	}

	errVars := []error{}
	for _, p := range publishers {
		err := p.Publish(ctx, e)
		if err != nil {
			errVars = append(errVars, err)
		}
	}

	if len(errVars) != 0 {
		return errors.Join(errVars...)
	}

	return nil
}
