package stockmarket

import (
	"context"
	"sync"
	"time"

	"github.com/nullableocean/grpcservices/order/domain"
	"go.uber.org/zap"
)

type EventStore interface {
	Save(ctx context.Context, event *domain.UpdateEvent) error
}

type StockMarketClient interface {
	CreateMarketOrder(ctx context.Context, o *domain.Order) error
}

type StockMarketEventBroker interface {
	GetEventsChan(ctx context.Context) (<-chan *domain.MarketEvent, error)
}

type Sub struct {
	Id   int
	UpCh <-chan *domain.UpdateEvent
}

type innersub struct {
	upCh chan<- *domain.UpdateEvent
}

type StockMarketService struct {
	client      StockMarketClient
	eventBroker StockMarketEventBroker
	eventStore  EventStore

	updatesSub map[int]*innersub
	nextSubId  int

	mu   sync.RWMutex
	stop chan struct{}
	wg   sync.WaitGroup

	logger *zap.Logger
}

func NewStockMarketService(logger *zap.Logger, stockMarketApi StockMarketClient, eventBroker StockMarketEventBroker, eventStore EventStore) (*StockMarketService, error) {
	s := &StockMarketService{
		client:      stockMarketApi,
		eventStore:  eventStore,
		eventBroker: eventBroker,

		mu:   sync.RWMutex{},
		stop: make(chan struct{}),
		wg:   sync.WaitGroup{},

		updatesSub: make(map[int]*innersub),
		logger:     logger,
	}

	err := s.ObserveEvents()
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (sm *StockMarketService) SendOrder(ctx context.Context, o *domain.Order) error {
	sm.logger.Info("send order on stock market", zap.Int64("order_id", o.Id()))

	err := sm.client.CreateMarketOrder(ctx, o)
	if err != nil {
		sm.logger.Warn("error send order on stock market", zap.Int64("order_id", o.Id()), zap.Error(err))

		return err
	}

	return nil
}

func (sm *StockMarketService) Updates() *Sub {
	updates := make(chan *domain.UpdateEvent)
	sub := &Sub{
		UpCh: updates,
	}

	sm.mu.Lock()

	sm.nextSubId++

	sub.Id = sm.nextSubId
	sm.updatesSub[sub.Id] = &innersub{upCh: updates}

	sm.mu.Unlock()

	return sub
}

func (sm *StockMarketService) sendUpdate(ctx context.Context, up *domain.UpdateEvent) {
	sm.mu.RLock()
SUB_LOOP:
	for _, sub := range sm.updatesSub {
		select {
		case <-ctx.Done():
			break SUB_LOOP
		case sub.upCh <- up:
		}
	}
	sm.mu.RUnlock()
	sm.wg.Done()
}

func (sm *StockMarketService) ObserveEvents() error {
	events, err := sm.eventBroker.GetEventsChan(context.Background())
	if err != nil {
		return err
	}

	go func(evs <-chan *domain.MarketEvent) {

		ctx, cl := context.WithCancel(context.Background())
		defer cl()
	EV_LOOP:
		for {
			select {
			case <-sm.stop:
				break EV_LOOP
			case e, ok := <-evs:
				if !ok {
					sm.logger.Info("market events channel was closed")

					break EV_LOOP
				}

				upEvent, err := sm.handleMarketEvent(e)
				if err != nil {
					sm.logger.Warn("error handle market event", zap.Int64("event_id", e.Id), zap.Error(err))
				}

				sm.wg.Add(1)
				sm.sendUpdate(ctx, upEvent)
			}
		}
	}(events)

	return nil
}

func (sm *StockMarketService) handleMarketEvent(mEvent *domain.MarketEvent) (*domain.UpdateEvent, error) {
	newUpdateEvent := &domain.UpdateEvent{
		OrderId:   mEvent.Id,
		NewStatus: mEvent.NewStatus,
		CreatedAt: time.Now(),
	}

	err := sm.eventStore.Save(context.Background(), newUpdateEvent)
	if err != nil {
		return nil, err
	}

	return newUpdateEvent, nil
}

func (sm *StockMarketService) Stop() {
	close(sm.stop)
	sm.wg.Wait()

	for _, sub := range sm.updatesSub {
		close(sub.upCh)
	}
}
