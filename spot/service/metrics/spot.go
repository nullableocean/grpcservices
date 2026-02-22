package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	Namespace             string = "spot"
	ViewMarketCounterName string = "view_markets_call_count"
	ViewMarketDuration    string = "view_markets_duration"
)

type SpotMetrics struct {
	viewMarketCounter      prometheus.Counter
	viewMarketExecuteTimer prometheus.Histogram
}

func NewSpotMetrics(registry *prometheus.Registry) *SpotMetrics {
	promFactory := promauto.With(registry)

	return &SpotMetrics{
		viewMarketCounter: promFactory.NewCounter(
			prometheus.CounterOpts{
				Namespace: Namespace,
				Name:      ViewMarketCounterName,
				Help:      "Total calls view market",
			}),
		viewMarketExecuteTimer: promFactory.NewHistogram(
			prometheus.HistogramOpts{
				Namespace: Namespace,
				Name:      ViewMarketDuration,
				Help:      "Duration for call view market",
			}),
	}
}

func (metrics *SpotMetrics) CalledViewMarket(duration time.Duration) {
	metrics.viewMarketCounter.Inc()
	metrics.viewMarketExecuteTimer.Observe(duration.Seconds())
}
