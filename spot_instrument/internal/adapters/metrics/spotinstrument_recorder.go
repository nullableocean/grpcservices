package metrics

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type SpotInstrumentRecorder struct {
	viewMarkets       prometheus.Counter
	failedViewMarkets prometheus.Counter
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
	}
}

func (r *SpotInstrumentRecorder) ViewMarkets(ctx context.Context) {
	r.viewMarkets.Inc()
}

func (r *SpotInstrumentRecorder) FailedViewMarkets(ctx context.Context) {
	r.failedViewMarkets.Inc()
}
