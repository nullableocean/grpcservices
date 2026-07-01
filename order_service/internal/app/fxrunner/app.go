package fxrunner

import (
	"fmt"

	"github.com/nullableocean/grpcservices/orderservice/internal/app/fxrunner/modules"
	"go.uber.org/fx"
)

func FxAppRunner() (*fx.App, error) {
	app := fx.New(
		modules.ConfigModule(),
		modules.LoggerModule(),
		modules.PostgresModule(),
		modules.RedisModule(),
		modules.KafkaProducerModule(),
		modules.MetricsModule(),
		modules.TelemetryModule(),
		modules.GrpcClientModule(),
		modules.SpotClientModule(),
		modules.RepositoriesModule(),
		modules.ServicesModule(),
		modules.EventPublishersModule(),
		modules.UpdateNotifierModule(),
		modules.GRPCServerModule(),
		modules.GRPCRunnerModule(),
		modules.RedisSubsriberModule(),
		modules.OutboxModule(),
		modules.HealthcheckModule(),
		modules.HTTPServerModule(),
	)
	if err := app.Err(); err != nil {
		return nil, fmt.Errorf("dependency graph error: %w", err)
	}

	return app, nil
}
