package stockmarket

import (
	"context"
	"math/rand/v2"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nullableocean/grpcservices/order/domain"
	"github.com/nullableocean/grpcservices/order/service/store/ram"
	"github.com/nullableocean/grpcservices/pkg/order"
)

type DummyMarketBroker struct {
	updatesq chan *domain.MarketEvent
	once     sync.Once

	orderStore *ram.OrderStore
	id         atomic.Int64
}

func NewDummyBroker(orderStore *ram.OrderStore) *DummyMarketBroker {
	b := &DummyMarketBroker{
		once:       sync.Once{},
		orderStore: orderStore,
		id:         atomic.Int64{},
	}

	b.once.Do(func() {
		b.updatesq = make(chan *domain.MarketEvent)
	})

	b.generateEvents()
	return b
}

func (b *DummyMarketBroker) GetEventsChan(ctx context.Context) (<-chan *domain.MarketEvent, error) {
	b.once.Do(func() {
		b.updatesq = make(chan *domain.MarketEvent)
	})

	return b.updatesq, nil
}

func (b *DummyMarketBroker) generateEvents() {
	ticker := time.NewTicker(time.Second * 10)

	go func() {
		for range ticker.C {
			orders := b.orderStore.GetAll(context.Background())
			b.genEvent(orders)
		}
	}()
}

func (b *DummyMarketBroker) genEvent(orders []*domain.Order) {
	if len(orders) == 0 {
		return
	}

	n := rand.IntN(len(orders))
	evOrder := orders[n]

	var newStatus order.OrderStatus
	if evOrder.Status() == order.ORDER_STATUS_CREATED {
		chance := rand.IntN(100)
		if chance < 21 {
			newStatus = order.ORDER_STATUS_REJECTED
		} else {
			newStatus = order.ORDER_STATUS_PENDING
		}
	} else {
		newStatus = order.ORDER_STATUS_COMPLETED
	}

	event := &domain.MarketEvent{
		Id:        b.id.Add(1),
		OrderId:   evOrder.Id(),
		NewStatus: newStatus,
		CreatedAt: time.Now(),
	}

	b.updatesq <- event
}
