package tests

import (
	"context"
	"testing"

	"github.com/nullableocean/grpcservices/api/spotpb"
	"github.com/nullableocean/grpcservices/spot/server"
	"github.com/nullableocean/grpcservices/spot/service/metrics"
	"github.com/nullableocean/grpcservices/spot/service/spot"
	"github.com/nullableocean/grpcservices/spot/service/store/ram"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestMetricsCollects(t *testing.T) {

	store := ram.NewMarketStore()
	spotService := spot.NewSpotInstrument(store)
	logger := zap.NewNop()

	reg := prometheus.NewRegistry()
	spotMetrics := metrics.NewSpotMetrics(reg)
	spotServer := server.NewSpotInstrumentServer(spotService, logger, spotMetrics)

	spotServer.ViewMarkets(context.Background(), &spotpb.ViewMarketsRequest{})
	spotServer.ViewMarkets(context.Background(), &spotpb.ViewMarketsRequest{})
	spotServer.ViewMarkets(context.Background(), &spotpb.ViewMarketsRequest{})

	gotMetricsFamily, err := reg.Gather()
	require.NoError(t, err)

	foundCounter := false
	foundCallDuration := false
	for _, m := range gotMetricsFamily {
		if *m.Name == metrics.Namespace+"_"+metrics.ViewMarketCounterName {
			foundCounter = true

			collectedMetrics := m.GetMetric()
			require.Len(t, collectedMetrics, 1)

			counterMetric := collectedMetrics[0]

			callCount := counterMetric.Counter.GetValue()
			require.Equal(t, int(callCount), 3, "call count should be 3")
		}

		if *m.Name == metrics.Namespace+"_"+metrics.ViewMarketDuration {
			foundCallDuration = true
		}
	}

	assert.True(t, foundCounter, "viewmarket calls counter metrics should found in collected metrics")
	assert.True(t, foundCallDuration, "viewmarket duration metrics should found in collected metrics")
}
