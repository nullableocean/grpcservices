package limiter

import (
	"context"
)

type Limiter struct {
	tokens chan struct{}
}

func New(size int) *Limiter {
	if size <= 0 {
		size = 1
	}

	return &Limiter{
		tokens: make(chan struct{}, size),
	}
}

func (l *Limiter) Acquire() {
	l.tokens <- struct{}{}
}

func (l *Limiter) Release() {
	select {
	case <-l.tokens:
	default:
	}
}

func (l *Limiter) AcquireContext(ctx context.Context) error {
	select {
	case l.tokens <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
