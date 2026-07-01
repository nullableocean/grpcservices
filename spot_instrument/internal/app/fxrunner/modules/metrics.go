package modules

import (
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/fx"
)

func MetricsModule() fx.Option {
	return fx.Options(
		fx.Provide(
			func() (*prometheus.Registry, *grpc_prometheus.ServerMetrics, *grpc_prometheus.ClientMetrics) {
				reg := prometheus.NewRegistry()
				srvMetrics := grpc_prometheus.NewServerMetrics()
				cliMetrics := grpc_prometheus.NewClientMetrics()
				reg.MustRegister(srvMetrics, cliMetrics)

				return reg, srvMetrics, cliMetrics
			},
		),
	)
}
