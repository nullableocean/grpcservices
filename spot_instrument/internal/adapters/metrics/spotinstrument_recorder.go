package metrics

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type SpotInstrumentRecorder struct {
	viewMarkets       prometheus.Counter
	failedViewMarkets prometheus.Counter
	failedFindMarket  prometheus.Counter
}

func NewSpotInstrumentRecorder(registry *prometheus.Registry) *SpotInstrumentRecorder {
	return &SpotInstrumentRecorder{
		viewMarkets: promauto.With(registry).NewCounter(prometheus.CounterOpts{
			Name: "view_markets_total",
			Help: "Total number of view markets calls",
		}),
		failedViewMarkets: promauto.With(registry).NewCounter(prometheus.CounterOpts{
			Name: "failed_view_markets_total",
			Help: "Total number of failed view markets calls",
		}),
		failedFindMarket: promauto.With(registry).NewCounter(prometheus.CounterOpts{
			Name: "failed_find_markets_total",
			Help: "Total number of failed find markets calls",
		}),
	}
}

func (r *SpotInstrumentRecorder) ViewMarkets(ctx context.Context) {
	r.viewMarkets.Inc()
}

func (r *SpotInstrumentRecorder) FailedViewMarkets(ctx context.Context) {
	r.failedViewMarkets.Inc()
}

func (r *SpotInstrumentRecorder) FailedFindMarket(ctx context.Context) {
	r.failedFindMarket.Inc()
}
