package modules

import (
	"context"
	"fmt"

	"github.com/nullableocean/grpcservices/orderservice/internal/config"
	"github.com/nullableocean/grpcservices/shared/telemetry"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func TelemetryModule() fx.Option {
	return fx.Options(
		fx.Provide(
			func(cfg *config.Config, logger *zap.Logger) (telemetry.ShutdownFunc, error) {
				cfgTelemetry := &telemetry.Config{
					ServiceName:        cfg.App.Name,
					ExporterGRPCAddr:   cfg.Telemetry.ExporterGrpcAddress,
					RatioSampler:       cfg.Telemetry.SampleRatio,
					BatchTimeout:       cfg.Telemetry.BatchTimeout,
					MaxExportBatchSize: cfg.Telemetry.MaxExportBatchSize,
					MaxQueueSize:       cfg.Telemetry.MaxQueueSize,
				}
				shutdown, err := telemetry.SetupGlobal(context.Background(), cfgTelemetry)
				if err != nil {
					return nil, fmt.Errorf("init telemetry: %w", err)
				}

				return shutdown, nil
			},
		),
		fx.Invoke(
			func(lc fx.Lifecycle, shutdown telemetry.ShutdownFunc) {
				lc.Append(fx.Hook{
					OnStop: func(ctx context.Context) error {
						return shutdown(ctx)
					},
				})
			},
		),
	)
}
