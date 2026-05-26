package retry

import (
	"math"
	"math/rand"
	"time"
)

// NewExponentialBackoffFunc возвращает экпоненциальню бэкофф функцию с полным jitter
func NewExponentialBackoffFunc(baseDelay, maxDelay time.Duration) func(attempt int) time.Duration {
	return func(attempt int) time.Duration {
		delay := baseDelay * time.Duration(math.Pow(2, float64(attempt)))
		if delay > maxDelay {
			delay = maxDelay
		}

		delay = time.Duration(rand.Int63n(int64(delay)))

		return delay
	}
}
