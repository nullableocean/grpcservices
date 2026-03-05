package streamer

import (
	"context"
	"sync"
	"time"

	"github.com/nullableocean/grpcservices/shared/limiter"
	"go.uber.org/zap"
)

var (
	defaultMaxSendingProcess = 5
	sendTimeout              = 5 * time.Second
	maxRetrySend             = 5

	subChannelBuf = 10
)

type Sub struct {
	Id       int
	ChangeCh <-chan Changes // канал только для чтения
}

type innersub struct {
	subId     int
	changesCh chan Changes
	close     chan struct{}
	sendMu    sync.RWMutex
	closeOnce sync.Once

	timeoutCount int
}

type ChangesStreamer struct {
	updatesSub    map[string]map[int]*innersub // orderUuid -> subId -> канал
	nextSubId     int
	mu            sync.RWMutex
	processLimits *limiter.Limiter

	logger *zap.Logger
}

type Option struct {
	maxSendingProcess int
}

func NewChangeStreamer(logger *zap.Logger, opt Option) *ChangesStreamer {
	if opt.maxSendingProcess <= 0 {
		opt.maxSendingProcess = defaultMaxSendingProcess
	}

	notifier := &ChangesStreamer{
		updatesSub:    make(map[string]map[int]*innersub),
		nextSubId:     0,
		mu:            sync.RWMutex{},
		processLimits: limiter.New(opt.maxSendingProcess),

		logger: logger,
	}

	return notifier
}

func (s *ChangesStreamer) Send(ctx context.Context, change Changes) error {
	if change.IsFinal() {
		return s.sendAndCloseSubs(ctx, change)
	}

	err := s.processLimits.AcquireContext(ctx)
	if err != nil {
		s.logger.Warn("send changes on streams cancelled by context", zap.Error(ctx.Err()))
		return err
	}

	go func() {
		defer s.processLimits.Release()
		s.notifySubs(change)
	}()

	return nil
}

func (s *ChangesStreamer) sendAndCloseSubs(ctx context.Context, change Changes) error {
	err := s.processLimits.AcquireContext(ctx)
	if err != nil {
		s.logger.Warn("send changes on streams cancelled by context", zap.Error(ctx.Err()))
		return err
	}

	go func() {
		defer s.processLimits.Release()

		s.notifySubs(change)
		s.closeSubs(change.ChangedOrder())
	}()

	return nil
}

func (s *ChangesStreamer) Sub(ctx context.Context, orderUuid string) (*Sub, error) {
	ch := make(chan Changes, subChannelBuf)
	closeCh := make(chan struct{})

	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextSubId++
	subId := s.nextSubId

	if s.updatesSub[orderUuid] == nil {
		s.updatesSub[orderUuid] = make(map[int]*innersub)
	}
	s.updatesSub[orderUuid][subId] = &innersub{
		subId:     subId,
		changesCh: ch,
		close:     closeCh,
	}

	return &Sub{
		Id:       subId,
		ChangeCh: ch,
	}, nil
}

func (s *ChangesStreamer) Dissub(ctx context.Context, orderUuid string, subId int) {
	s.mu.Lock()
	subs, ok := s.updatesSub[orderUuid]
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
		delete(s.updatesSub, orderUuid)
	}
	s.mu.Unlock()

	sub.closeOnce.Do(func() {
		close(sub.close)

		sub.sendMu.Lock()
		defer sub.sendMu.Unlock()
		close(sub.changesCh)
	})
}

func (s *ChangesStreamer) notifySubs(change Changes) {
	orderUuid := change.ChangedOrder()

	logger := s.logger.With(
		zap.String("order_uuid", orderUuid),
	)

	s.mu.RLock()
	subs, ok := s.updatesSub[orderUuid]
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

		select {
		case <-sub.close:
		case sub.changesCh <- change:
		case <-time.After(sendTimeout):
			logger.Warn("timeout sending update to subscriber", zap.Int("sub_id", sub.subId))

			sub.timeoutCount++
			if sub.timeoutCount > maxRetrySend {
				logger.Warn("limit retry send to subscriber", zap.Int("sub_id", sub.subId))

				go s.Dissub(context.Background(), orderUuid, sub.subId)
			}
		}

		sub.sendMu.RUnlock()
	}
}

func (s *ChangesStreamer) closeSubs(orderUuid string) {
	s.mu.Lock()
	subs, ok := s.updatesSub[orderUuid]
	if !ok {
		s.mu.Unlock()
		return
	}
	delete(s.updatesSub, orderUuid)
	s.mu.Unlock()

	for _, sub := range subs {
		sub.closeOnce.Do(func() {
			close(sub.close)
			sub.sendMu.Lock()
			defer sub.sendMu.Unlock()
			close(sub.changesCh)
		})
	}
}
