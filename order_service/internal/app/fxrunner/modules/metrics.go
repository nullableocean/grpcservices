package modules

import (
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/metrics"
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
		fx.Provide(func(reg *prometheus.Registry) *metrics.KafkaMetricsRecorder {
			return metrics.NewKafkaMetricsRecorder(reg)
		}),
		fx.Provide(func(reg *prometheus.Registry) *metrics.OrderMetricsRecorder {
			return metrics.NewOrderMetricsRecorder(reg)
		}),
		fx.Provide(func(reg *prometheus.Registry) *metrics.OutboxMetricsRecorder {
			return metrics.NewOutboxMetricsRecorder(reg)
		}),
		fx.Provide(func(reg *prometheus.Registry) *metrics.CircuitBreakerMetricsRecorder {
			return metrics.NewCircuitBreakerMetricsRecorder(reg)
		}),
		fx.Provide(func(reg *prometheus.Registry) *metrics.RedisMetricsRecorder {
			return metrics.NewRedisMetricsRecorder(reg)
		}),
	)
}
