package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	CIRCUIT_BREAKER_SUCCESS       = "success"
	CIRCUIT_BREAKER_FAILURE       = "failure"
	CIRCUIT_BREAKER_SHORT_CIRCUIT = "short_circuit"

	CIRCUIT_BREAKER_STATE_CLOSED    = 0
	CIRCUIT_BREAKER_STATE_OPEN      = 1
	CIRCUIT_BREAKER_STATE_HALF_OPEN = 2
)

type CircuitBreakerMetricsRecorder struct {
	requestsTotal       *prometheus.CounterVec
	stateTransitions    *prometheus.CounterVec
	currentState        prometheus.Gauge
	consecutiveFailures prometheus.Gauge
}

func NewCircuitBreakerMetricsRecorder(registry *prometheus.Registry) *CircuitBreakerMetricsRecorder {
	return &CircuitBreakerMetricsRecorder{
		requestsTotal: promauto.With(registry).NewCounterVec(prometheus.CounterOpts{
			Name: "circuit_breaker_requests_total",
			Help: "Total number of requests processed by circuit breaker",
		}, []string{"result"}),
		stateTransitions: promauto.With(registry).NewCounterVec(prometheus.CounterOpts{
			Name: "circuit_breaker_state_transitions_total",
			Help: "Number of state transitions",
		}, []string{"from", "to"}),
		currentState: promauto.With(registry).NewGauge(prometheus.GaugeOpts{
			Name: "circuit_breaker_current_state",
			Help: "Current state: 0=closed, 1=open, 2=half-open",
		}),
		consecutiveFailures: promauto.With(registry).NewGauge(prometheus.GaugeOpts{
			Name: "circuit_breaker_consecutive_failures",
			Help: "Current number of consecutive failures",
		}),
	}
}

func (m *CircuitBreakerMetricsRecorder) RecordRejectedRequest() {
	m.requestsTotal.WithLabelValues(CIRCUIT_BREAKER_SHORT_CIRCUIT).Inc()
}

func (m *CircuitBreakerMetricsRecorder) RecordSuccessfulRequest() {
	m.requestsTotal.WithLabelValues(CIRCUIT_BREAKER_SUCCESS).Inc()
}

func (m *CircuitBreakerMetricsRecorder) RecordFailedRequest() {
	m.requestsTotal.WithLabelValues(CIRCUIT_BREAKER_FAILURE).Inc()
}

// RecordStateChange увеличивает счётчик переходов между состояниями.
func (m *CircuitBreakerMetricsRecorder) RecordStateTransition(from, to string) {
	m.stateTransitions.WithLabelValues(from, to).Inc()
}

// SetConsecutiveFailures устанавливает количество последовательных ошибок.
func (m *CircuitBreakerMetricsRecorder) SetConsecutiveFailures(failures float64) {
	m.consecutiveFailures.Set(failures)
}

// SetClosedState устанавливают текущее состояние для gauge.
func (m *CircuitBreakerMetricsRecorder) SetClosedState() {
	m.currentState.Set(CIRCUIT_BREAKER_STATE_CLOSED)
}

// SetOpenState устанавливают текущее состояние для gauge.
func (m *CircuitBreakerMetricsRecorder) SetOpenState() {
	m.currentState.Set(CIRCUIT_BREAKER_STATE_OPEN)
}

// SetHalfOpenState устанавливают текущее состояние для gauge.
func (m *CircuitBreakerMetricsRecorder) SetHalfOpenState() {
	m.currentState.Set(CIRCUIT_BREAKER_STATE_HALF_OPEN)
}
