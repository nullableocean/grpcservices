package metrics

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type PrometheusMetricsRecorder struct {
	orderCreated      prometheus.Counter
	orderCompleted    prometheus.Counter
	orderRejected     prometheus.Counter
	orderCancelled    prometheus.Counter
	orderCreateFailed prometheus.Counter
	orderUpdateFailed prometheus.Counter
}

func NewPrometheusMetricsRecorder(registry *prometheus.Registry) *PrometheusMetricsRecorder {
	m := &PrometheusMetricsRecorder{
		orderCreated: promauto.With(registry).NewCounter(prometheus.CounterOpts{
			Name: "orders_created_total",
			Help: "Total number of orders created",
		}),
		orderCompleted: promauto.With(registry).NewCounter(prometheus.CounterOpts{
			Name: "orders_completed_total",
			Help: "Total number of orders completed",
		}),
		orderRejected: promauto.With(registry).NewCounter(prometheus.CounterOpts{
			Name: "orders_rejected_total",
			Help: "Total number of orders rejected",
		}),
		orderCancelled: promauto.With(registry).NewCounter(prometheus.CounterOpts{
			Name: "orders_cancelled_total",
			Help: "Total number of orders cancelled",
		}),
		orderCreateFailed: promauto.With(registry).NewCounter(prometheus.CounterOpts{
			Name: "orders_create_failed_total",
			Help: "Total number of creating orders failed",
		}),
		orderUpdateFailed: promauto.With(registry).NewCounter(prometheus.CounterOpts{
			Name: "orders_update_failed_total",
			Help: "Total number of update order failed",
		}),
	}

	return m
}

func (m *PrometheusMetricsRecorder) OrderCreated(ctx context.Context) {
	m.orderCreated.Inc()
}

func (m *PrometheusMetricsRecorder) OrderCompleted(ctx context.Context) {
	m.orderCompleted.Inc()
}

func (m *PrometheusMetricsRecorder) OrderRejected(ctx context.Context) {
	m.orderRejected.Inc()
}

func (m *PrometheusMetricsRecorder) OrderCancelled(ctx context.Context) {
	m.orderCancelled.Inc()
}

func (m *PrometheusMetricsRecorder) OrderFailedCreate(ctx context.Context) {
	m.orderCreateFailed.Inc()
}

func (m *PrometheusMetricsRecorder) OrderFailedUpdate(ctx context.Context) {
	m.orderUpdateFailed.Inc()
}
