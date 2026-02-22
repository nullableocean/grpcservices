package metrics

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	Namespace           string = "order"
	CreateOrderCalls    string = "create_order_call_count"
	CreateOrderDuration string = "create_order_call_duration"
	GetStatusCalls      string = "get_order_status_call_count"
)

type OrderServiceMetrics struct {
	getStatusCounter    *prometheus.CounterVec
	createOrderCounter  *prometheus.CounterVec
	createOrderDuration *prometheus.HistogramVec
}

func NewOrderMetrics(registry *prometheus.Registry) *OrderServiceMetrics {
	promFactory := promauto.With(registry)

	return &OrderServiceMetrics{
		getStatusCounter: promFactory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: Namespace,
				Name:      GetStatusCalls,
				Help:      "Total calls get status for order",
			}, []string{"user", "order"}),
		createOrderCounter: promFactory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: Namespace,
				Name:      CreateOrderCalls,
				Help:      "Total calls create order",
			}, []string{"user"}),
		createOrderDuration: promFactory.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: Namespace,
				Name:      CreateOrderDuration,
				Help:      "Duration for call create order",
			}, []string{"user"}),
	}
}

func (metrics *OrderServiceMetrics) CalledCreateOrder(userId int64, duration time.Duration) {
	userIdStr := strconv.FormatInt(userId, 10)
	metrics.createOrderCounter.WithLabelValues(userIdStr).Inc()
	metrics.createOrderDuration.WithLabelValues(userIdStr).Observe(duration.Seconds())
}

func (metrics *OrderServiceMetrics) CalledGetStatus(userId, orderId int64) {
	userIdStr := strconv.FormatInt(userId, 10)
	orderIdStr := strconv.FormatInt(orderId, 10)

	metrics.getStatusCounter.WithLabelValues(userIdStr, orderIdStr).Inc()
}
