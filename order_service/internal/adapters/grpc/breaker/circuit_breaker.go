package breaker

import (
	"sync"

	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/grpc/mapping"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/metrics"
	"github.com/sony/gobreaker"
	"go.uber.org/zap"
)

type CircuitBreaker struct {
	cb      *gobreaker.CircuitBreaker
	logger  *zap.Logger
	metrics *metrics.CircuitBreakerMetricsRecorder

	stateMu      sync.Mutex
	lastState    gobreaker.State
	lastFailures uint32
}

func NewCircuitBreaker(logger *zap.Logger, metrics *metrics.CircuitBreakerMetricsRecorder, cb *gobreaker.CircuitBreaker) *CircuitBreaker {
	breaker := &CircuitBreaker{
		cb:           cb,
		logger:       logger,
		metrics:      metrics,
		stateMu:      sync.Mutex{},
		lastState:    gobreaker.StateClosed,
		lastFailures: 0,
	}

	metrics.SetClosedState()
	metrics.SetConsecutiveFailures(0)

	return breaker
}

func (cb *CircuitBreaker) Execute(fn func() (interface{}, error)) (interface{}, error) {
	cb.updateStateAndMetrics()
	defer cb.updateStateAndMetrics()

	r, err := cb.cb.Execute(func() (interface{}, error) {
		r, err := fn()

		if mapping.IsClientSideGrpcError(err) {
			cb.logger.Debug("circuit breaker skip client side error", zap.Error(err))

			return r, nil
		}

		return r, err
	})

	if err != nil {
		if err == gobreaker.ErrOpenState {
			cb.logger.Warn("circuit breaker is open, request rejected")
			cb.metrics.RecordRejectedRequest()

			return nil, err
		}

		cb.logger.Error("circuit breaker execution failed", zap.Error(err))
		cb.metrics.RecordFailedRequest()

		return nil, err
	}

	cb.metrics.RecordSuccessfulRequest()

	return r, nil
}

func (cb *CircuitBreaker) updateStateAndMetrics() {
	state := cb.cb.State()
	counts := cb.cb.Counts()

	cb.metrics.SetConsecutiveFailures(float64(counts.ConsecutiveFailures))
	cb.recordState(state)

	cb.stateMu.Lock()
	defer cb.stateMu.Unlock()

	if cb.lastState != state {
		fromStr := cb.stateToString(cb.lastState)
		toStr := cb.stateToString(state)

		cb.metrics.RecordStateTransition(fromStr, toStr)

		cb.logger.Debug("circuit breaker state changed",
			zap.String("from", fromStr),
			zap.String("to", toStr),
			zap.Uint32("consecutive_failures", counts.ConsecutiveFailures),
			zap.Uint32("total_failures", counts.TotalFailures),
			zap.Uint32("total_successes", counts.TotalSuccesses),
		)

		cb.lastState = state
	}
}

func (cb *CircuitBreaker) recordState(s gobreaker.State) {
	switch s {
	case gobreaker.StateClosed:
		cb.metrics.SetClosedState()
	case gobreaker.StateOpen:
		cb.metrics.SetOpenState()
	case gobreaker.StateHalfOpen:
		cb.metrics.SetHalfOpenState()
	}
}

func (cb *CircuitBreaker) stateToString(s gobreaker.State) string {
	switch s {
	case gobreaker.StateClosed:
		return "closed"
	case gobreaker.StateOpen:
		return "open"
	case gobreaker.StateHalfOpen:
		return "half_open"
	default:
		return "unknown"
	}
}
