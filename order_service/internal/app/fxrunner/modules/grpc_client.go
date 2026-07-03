package modules

import (
	"context"
	"fmt"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	spotv1 "github.com/nullableocean/grpcservices/api/gen/spot/v1"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/grpc/breaker"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/grpc/client"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/metrics"
	"github.com/nullableocean/grpcservices/orderservice/internal/config"
	shared_inters "github.com/nullableocean/grpcservices/shared/interceptors"
	"github.com/nullableocean/grpcservices/shared/logger"
	"github.com/sony/gobreaker"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
)

func GrpcClientModule() fx.Option {
	return fx.Options(
		fx.Provide(func(ctxZapLogger *logger.CtxZapLogger, zapLogger *zap.Logger, cfg *config.Config, clientMetrics *grpc_prometheus.ClientMetrics, brMetrics *metrics.CircuitBreakerMetricsRecorder) (*grpc.ClientConn, error) {
			cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
				Name:        "spotinstrument-breaker",
				MaxRequests: cfg.CircuitBreaker.MaxRequests,
				Interval:    cfg.CircuitBreaker.Interval,
				Timeout:     cfg.CircuitBreaker.Timeout,
			})
			cbWithMetrics := breaker.NewCircuitBreaker(ctxZapLogger, brMetrics, cb)

			retryOpts := []retry.CallOption{
				retry.WithMax(uint(cfg.Retry.MaxRetries)),
				retry.WithBackoff(retry.BackoffExponential(cfg.Retry.Backoff)),
				retry.WithCodes(codes.Unavailable, codes.DeadlineExceeded, codes.ResourceExhausted),
			}

			interceptors := grpc.WithChainUnaryInterceptor(
				shared_inters.UnaryClientPanicRecovery(zapLogger, cfg.Log.StackLines),
				shared_inters.UnaryClientXReqId(),
				shared_inters.UnaryClientXReqIdTelemetry(),
				clientMetrics.UnaryClientInterceptor(),
				shared_inters.UnaryClientLogger(zapLogger),
				shared_inters.UnaryClientJwtForwardInterceptor(),
				retry.UnaryClientInterceptor(retryOpts...),
				shared_inters.UnaryCircuitBreakerInterceptor(cbWithMetrics),
			)

			conn, err := grpc.NewClient(cfg.Spot.Endpoint,
				grpc.WithTransportCredentials(insecure.NewCredentials()),
				grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
				interceptors,
			)
			if err != nil {
				return nil, fmt.Errorf("failed create grpc client for spot: %w", err)
			}
			return conn, nil
		},
		),
		fx.Invoke(
			func(lc fx.Lifecycle, conn *grpc.ClientConn) {
				lc.Append(fx.Hook{
					OnStop: func(ctx context.Context) error {
						return conn.Close()
					},
				})
			},
		),
	)
}

func SpotClientModule() fx.Option {
	return fx.Options(
		fx.Provide(
			func(logger *logger.CtxZapLogger, conn *grpc.ClientConn, cfg *config.Config) (*client.SpotInstrumentClient, error) {
				return client.NewSpotInstrumentClient(logger, spotv1.NewSpotInstrumentClient(conn), client.Option{
					RequestTimeout: cfg.GRPC.ClientTimeout,
				})
			},
		),
	)
}
