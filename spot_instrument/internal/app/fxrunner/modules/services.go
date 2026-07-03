package modules

import (
	"github.com/nullableocean/grpcservices/shared/logger"
	"github.com/nullableocean/grpcservices/spotinstrument/internal/adapters/metrics"
	"github.com/nullableocean/grpcservices/spotinstrument/internal/adapters/repository/postgres"
	"github.com/nullableocean/grpcservices/spotinstrument/internal/core/services/spotinstrument"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/fx"
)

func ServicesModule() fx.Option {
	return fx.Options(
		fx.Provide(
			func(ctxLogger *logger.CtxZapLogger, repo *postgres.MarketRepository, reg *prometheus.Registry) *spotinstrument.SpotInstrument {
				metricsRecorder := metrics.NewSpotInstrumentRecorder(reg)
				return spotinstrument.NewSpotInstrument(ctxLogger, repo, metricsRecorder)
			},
		),
	)
}
