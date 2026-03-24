package updatenotifier

import (
	"sync"

	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
)

// var _ ports.Sub = &Sub{}

type Sub struct {
	id        int
	updatesCh chan *model.EventOrderUpdated

	closeCh   chan struct{}
	closeOnce sync.Once
	timeouts  int32
}

func (s *Sub) Close() {
	s.closeOnce.Do(func() {
		close(s.closeCh)
	})
}

func (s *Sub) Updates() <-chan *model.EventOrderUpdated {
	return s.updatesCh
}

type Subs struct {
	subs   map[int]*Sub
	mu     sync.Mutex
	nextId int
}

func NewSubs() *Subs {
	return &Subs{
		subs:   map[int]*Sub{},
		mu:     sync.Mutex{},
		nextId: 0,
	}
}

func (subs *Subs) Add(sub *Sub) {
	subs.mu.Lock()

	subs.nextId++
	sub.id = subs.nextId
	subs.subs[subs.nextId] = sub

	subs.mu.Unlock()
}

func (subs *Subs) Remove(id int) {
	subs.mu.Lock()

	sub, ex := subs.subs[id]
	if !ex {
		return
	}

	sub.Close()
	delete(subs.subs, id)

	subs.mu.Unlock()
}

func (subs *Subs) GetSubs() []*Sub {
	subs.mu.Lock()

	cpSubs := make([]*Sub, 0, len(subs.subs))
	for _, s := range subs.subs {
		cpSubs = append(cpSubs, s)
	}

	subs.mu.Unlock()

	return cpSubs
}
