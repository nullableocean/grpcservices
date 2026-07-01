package fxrunner

import (
	"fmt"

	"github.com/nullableocean/grpcservices/spotinstrument/internal/app/fxrunner/modules"
	"go.uber.org/fx"
)

func FxAppRunner() (*fx.App, error) {
	app := fx.New(
		modules.ConfigModule(),
		modules.LoggerModule(),
		modules.PostgresModule(),
		modules.MetricsModule(),
		modules.TelemetryModule(),
		modules.RepositoriesModule(),
		modules.ServicesModule(),
		modules.GRPCServerModule(),
		modules.GRPCRunnerModule(),
		modules.HealthModule(),
		modules.HTTPServerModule(),
	)
	if err := app.Err(); err != nil {
		return nil, fmt.Errorf("dependency graph error: %w", err)
	}

	return app, nil
}
