package metrics

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type OutboxMetricsRecorder struct {
	eventsFetched prometheus.Counter
	eventsRelayed prometheus.Counter
	eventsFailed  prometheus.Counter
}

func NewOutboxMetricsRecorder(registry *prometheus.Registry) *OutboxMetricsRecorder {
	return &OutboxMetricsRecorder{
		eventsFetched: promauto.With(registry).NewCounter(prometheus.CounterOpts{
			Name: "outbox_events_fetched_total",
			Help: "Total number of outbox events fetched for processing",
		}),
		eventsRelayed: promauto.With(registry).NewCounter(prometheus.CounterOpts{
			Name: "outbox_events_relayed_total",
			Help: "Total number of outbox events successfully relayed",
		}),
		eventsFailed: promauto.With(registry).NewCounter(prometheus.CounterOpts{
			Name: "outbox_events_failed_total",
			Help: "Total number of outbox events that failed relay",
		}),
	}
}

func (r *OutboxMetricsRecorder) EventFetched(ctx context.Context) {
	r.eventsFetched.Inc()
}

func (r *OutboxMetricsRecorder) EventRelayed(ctx context.Context) {
	r.eventsRelayed.Inc()
}

func (r *OutboxMetricsRecorder) EventFailed(ctx context.Context) {
	r.eventsFailed.Inc()
}
