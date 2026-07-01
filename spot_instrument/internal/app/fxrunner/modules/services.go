package modules

import (
	"github.com/nullableocean/grpcservices/spotinstrument/internal/adapters/metrics"
	"github.com/nullableocean/grpcservices/spotinstrument/internal/adapters/repository/postgres"
	"github.com/nullableocean/grpcservices/spotinstrument/internal/core/services/spotinstrument"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func ServicesModule() fx.Option {
	return fx.Options(
		fx.Provide(
			func(logger *zap.Logger, repo *postgres.MarketRepository, reg *prometheus.Registry) *spotinstrument.SpotInstrument {
				metricsRecorder := metrics.NewSpotInstrumentRecorder(reg)
				return spotinstrument.NewSpotInstrument(logger, repo, metricsRecorder)
			},
		),
	)
}
