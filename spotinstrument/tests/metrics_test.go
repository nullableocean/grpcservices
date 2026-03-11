package tests

import (
	"context"
	"testing"

	spotv1 "github.com/nullableocean/grpcservices/api/gen/spot/v1"
	"github.com/nullableocean/grpcservices/shared/eventbus"
	guard "github.com/nullableocean/grpcservices/spotinstrumentinstrument/internal/service/auth"
	"github.com/nullableocean/grpcservices/spotinstrumentinstrument/internal/service/metrics"
	"github.com/nullableocean/grpcservices/spotinstrumentinstrument/internal/service/spot"
	"github.com/nullableocean/grpcservices/spotinstrumentinstrument/internal/store/ram"
	"github.com/nullableocean/grpcservices/spotinstrumentinstrument/internal/transport/grpc/server"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestMetricsCollects(t *testing.T) {

	store := ram.NewMarketStore()
	roleInspector := guard.NewRoleInspector()
	spotService := spot.NewSpotInstrument(zap.NewNop(), store, roleInspector, eventbus.NewEventBus(zap.NewNop(), eventbus.Option{}))
	logger := zap.NewNop()

	reg := prometheus.NewRegistry()
	spotMetrics := metrics.NewSpotMetrics(reg)
	spotServer := server.NewSpotInstrumentServer(spotService, logger, spotMetrics)

	spotServer.ViewMarkets(context.Background(), &spotv1.ViewMarketsRequest{})
	spotServer.ViewMarkets(context.Background(), &spotv1.ViewMarketsRequest{})
	spotServer.ViewMarkets(context.Background(), &spotv1.ViewMarketsRequest{})

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
