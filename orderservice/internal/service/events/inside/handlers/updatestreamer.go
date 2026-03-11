package handlers

import (
	"context"
	"sync"
	"time"

	"github.com/nullableocean/grpcservices/orderservice/internal/service/events/inside"
	"github.com/nullableocean/grpcservices/shared/eventbus"
	"github.com/nullableocean/grpcservices/shared/limiter"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

const (
	defaultMaxSendingProcess = 5
	sendTimeout              = 5 * time.Second
	maxRetrySend             = 5
	subChannelBuf            = 10
)

type Sub struct {
	Id      int
	EventCh <-chan inside.NewStatusEvent
}

type innersub struct {
	subId        int
	eventCh      chan inside.NewStatusEvent
	close        chan struct{}
	sendMu       sync.RWMutex
	closeOnce    sync.Once
	timeoutCount int
}

type StatusStreamer struct {
	subsByOrder   map[string]map[int]*innersub // order_uuid → subId → подписчик
	nextSubId     int
	mu            sync.RWMutex
	processLimits *limiter.Limiter
	logger        *zap.Logger
}

type Option struct {
	MaxSendingProcess int
}

func NewStatusStreamer(logger *zap.Logger, opt Option) *StatusStreamer {
	if opt.MaxSendingProcess <= 0 {
		opt.MaxSendingProcess = defaultMaxSendingProcess
	}
	return &StatusStreamer{
		subsByOrder:   make(map[string]map[int]*innersub),
		nextSubId:     0,
		processLimits: limiter.New(opt.MaxSendingProcess),
		logger:        logger,
	}
}

func (s *StatusStreamer) Subscribe(ctx context.Context, orderUuid string) (*Sub, error) {
	ch := make(chan inside.NewStatusEvent, subChannelBuf)
	closeCh := make(chan struct{})

	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextSubId++
	subId := s.nextSubId

	if s.subsByOrder[orderUuid] == nil {
		s.subsByOrder[orderUuid] = make(map[int]*innersub)
	}
	s.subsByOrder[orderUuid][subId] = &innersub{
		subId:   subId,
		eventCh: ch,
		close:   closeCh,
	}

	return &Sub{
		Id:      subId,
		EventCh: ch,
	}, nil
}

func (s *StatusStreamer) Unsubscribe(ctx context.Context, orderUuid string, subId int) {
	s.mu.Lock()
	subs, ok := s.subsByOrder[orderUuid]
	if !ok {
		s.mu.Unlock()
		return
	}
	sub, ok := subs[subId]
	if !ok {
		s.mu.Unlock()
		return
	}
	delete(subs, subId)
	if len(subs) == 0 {
		delete(s.subsByOrder, orderUuid)
	}
	s.mu.Unlock()

	sub.closeOnce.Do(func() {
		close(sub.close)
		sub.sendMu.Lock()
		defer sub.sendMu.Unlock()
		close(sub.eventCh)
	})
}

func (s *StatusStreamer) Handle(ctx context.Context, e eventbus.Event) {
	ctx, span := otel.Tracer("stream_notifier").Start(ctx, "handle_update_event")
	defer span.End()

	statusEvent, ok := e.(*inside.NewStatusEvent)
	if !ok {
		return
	}

	if err := s.processLimits.AcquireContext(ctx); err != nil {
		s.logger.Warn("failed to acquire limit for status stream", zap.Error(err))
		return
	}

	ctx = context.WithoutCancel(ctx)
	if s.isFinalEvent(statusEvent) {
		s.logger.Info("send update to subsribers and close all", zap.String("order_uuid", statusEvent.OrderUuid))

		go s.dispatchAndClose(ctx, statusEvent)
		return
	}

	s.logger.Info("send update to subsribers", zap.String("order_uuid", statusEvent.OrderUuid))
	go s.dispatchToSubscribers(ctx, statusEvent)
}

func (s *StatusStreamer) dispatchAndClose(ctx context.Context, event *inside.NewStatusEvent) {
	ctx, span := otel.Tracer("stream_notifier").Start(ctx, "dispatch_to_subscribers_and_close")
	defer span.End()

	s.dispatchToSubscribers(ctx, event)
	s.closeOrderSubs(event.OrderUuid)
}

func (s *StatusStreamer) dispatchToSubscribers(ctx context.Context, event *inside.NewStatusEvent) {
	ctx, span := otel.Tracer("stream_notifier").Start(ctx, "dispatch_to_subscribers")
	defer span.End()

	defer s.processLimits.Release()
	defer s.handlePanic()

	orderUuid := event.OrderUuid
	logger := s.logger.With(zap.String("order_uuid", orderUuid))

	s.mu.RLock()
	subs, ok := s.subsByOrder[orderUuid]
	if !ok {
		s.mu.RUnlock()
		return
	}
	subList := make([]*innersub, 0, len(subs))
	for _, sub := range subs {
		subList = append(subList, sub)
	}
	s.mu.RUnlock()

	for _, sub := range subList {
		sub.sendMu.RLock()
		timer := time.NewTimer(sendTimeout)
		select {
		case <-sub.close:
			timer.Stop()
		case <-ctx.Done():
			timer.Stop()
			sub.sendMu.RUnlock()
			return
		case sub.eventCh <- *event:
			timer.Stop()
		case <-timer.C:
			logger.Warn("timeout sending event to subscriber", zap.Int("sub_id", sub.subId))
			sub.timeoutCount++
			if sub.timeoutCount > maxRetrySend {
				logger.Warn("max retry exceeded, removing subscriber", zap.Int("sub_id", sub.subId))
				go s.unsubscribe(orderUuid, sub.subId)
			}
		}
		sub.sendMu.RUnlock()
	}
}

func (s *StatusStreamer) unsubscribe(orderUuid string, subId int) {
	s.Unsubscribe(context.Background(), orderUuid, subId)
}

func (s *StatusStreamer) CloseAll() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for orderUuid, subs := range s.subsByOrder {
		for _, sub := range subs {
			sub.closeOnce.Do(func() {
				close(sub.close)
				sub.sendMu.Lock()
				close(sub.eventCh)
				sub.sendMu.Unlock()
			})
		}
		delete(s.subsByOrder, orderUuid)
	}
}

func (s *StatusStreamer) closeOrderSubs(orderUuid string) {
	s.mu.Lock()
	subs, ok := s.subsByOrder[orderUuid]
	if !ok {
		s.mu.Unlock()
		return
	}
	delete(s.subsByOrder, orderUuid)
	s.mu.Unlock()

	for _, sub := range subs {
		sub.closeOnce.Do(func() {
			close(sub.close)
			sub.sendMu.Lock()
			close(sub.eventCh)
			sub.sendMu.Unlock()
		})
	}
}

func (s *StatusStreamer) isFinalEvent(e *inside.NewStatusEvent) bool {
	return e.NewStatus.IsFinal()
}

func (s *StatusStreamer) handlePanic() {
	if r := recover(); r != nil {
		s.logger.Error("panic in stream status goroutine", zap.Any("recover", r), zap.Stack("stack"))
	}
}
