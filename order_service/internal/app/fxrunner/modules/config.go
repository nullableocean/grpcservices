package modules

import (
	"github.com/nullableocean/grpcservices/orderservice/internal/config"
	"go.uber.org/fx"
)

func ConfigModule() fx.Option {
	return fx.Options(
		fx.Provide(config.NewConfig),
	)
}
