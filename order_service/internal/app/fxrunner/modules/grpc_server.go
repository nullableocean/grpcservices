package modules

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync/atomic"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	orderv1 "github.com/nullableocean/grpcservices/api/gen/order/v1"
	updatenotifier "github.com/nullableocean/grpcservices/orderservice/internal/adapters/events/update_notifier"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/grpc/server"
	"github.com/nullableocean/grpcservices/orderservice/internal/config"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/services/order"
	shared_auth "github.com/nullableocean/grpcservices/shared/auth"
	"github.com/nullableocean/grpcservices/shared/health"
	shared_inters "github.com/nullableocean/grpcservices/shared/interceptors"
	"github.com/nullableocean/grpcservices/shared/logger"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

func GRPCServerModule() fx.Option {
	return fx.Options(
		fx.Provide(
			func(logger *zap.Logger, cfg *config.Config, serverMetrics *grpc_prometheus.ServerMetrics) (*grpc.Server, error) {
				jwtParser := shared_auth.NewHmacJwtParser(cfg.Auth.JWTSecret)

				unaryInterceptors := grpc.ChainUnaryInterceptor(
					shared_inters.UnaryServerPanicRecovery(logger, cfg.Log.StackLines),
					shared_inters.UnaryJwtAuthInterceptor(logger, jwtParser),
					shared_inters.UnaryServerTelemetry(),
					shared_inters.UnaryServerLogger(logger),
					serverMetrics.UnaryServerInterceptor(),
					shared_inters.ValidationUnaryInterceptor(logger),
				)

				streamInterceptors := grpc.ChainStreamInterceptor(
					shared_inters.StreamServerPanicRecovery(logger, cfg.Log.StackLines),
					serverMetrics.StreamServerInterceptor(),
					shared_inters.StreamJwtAuthInterceptor(logger, jwtParser),
				)

				opts := []grpc.ServerOption{
					grpc.StatsHandler(otelgrpc.NewServerHandler()),
					grpc.KeepaliveParams(keepalive.ServerParameters{
						Time:    cfg.GRPC.Keepalive.Time,
						Timeout: cfg.GRPC.Keepalive.Timeout,
					}),
					grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
						PermitWithoutStream: cfg.GRPC.Keepalive.PermitWithoutStream,
					}),
					grpc.MaxRecvMsgSize(cfg.GRPC.ServerMaxRecvMsgSize),
					grpc.MaxSendMsgSize(cfg.GRPC.ServerMaxSendMsgSize),
					grpc.MaxConcurrentStreams(cfg.GRPC.ServerMaxConcurrentStreams),
					unaryInterceptors,
					streamInterceptors,
				}

				return grpc.NewServer(opts...), nil
			},
		),
		fx.Invoke(
			func(lc fx.Lifecycle, logger *zap.Logger, server *grpc.Server) {
				lc.Append(fx.Hook{
					OnStop: func(ctx context.Context) error {
						logger.Info("stopping gRPC server...")
						server.GracefulStop()
						return nil
					},
				})
			},
		),
	)
}

type GrpcServerHealth struct {
	ready atomic.Bool
}

func (o *GrpcServerHealth) HealthCheck() health.HealthCheck {
	return health.NewHealthCheck("grpc_server", func(ctx context.Context) error {
		if !o.ready.Load() {
			return errors.New("grpc server not ready")
		}
		return nil
	})
}

func GRPCRunnerModule() fx.Option {
	return fx.Options(
		fx.Provide(func() *GrpcServerHealth {
			return &GrpcServerHealth{}
		}),
		fx.Invoke(
			func(lc fx.Lifecycle, ctxLogger *logger.CtxZapLogger, logger *zap.Logger, cfg *config.Config, grpcServer *grpc.Server, orderService *order.OrderService, notifier *updatenotifier.UpdateNotifier, health *GrpcServerHealth) {
				orderServer := server.NewOrderServer(ctxLogger, orderService, notifier)
				orderv1.RegisterOrderServer(grpcServer, orderServer)

				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						lis, err := net.Listen("tcp", net.JoinHostPort(cfg.App.Address, cfg.App.Port))
						if err != nil {
							return fmt.Errorf("create listener: %w", err)
						}

						health.ready.Store(true)
						go func() {
							logger.Info("gRPC server started", zap.String("addr", lis.Addr().String()))
							if err := grpcServer.Serve(lis); err != nil {
								logger.Error("gRPC server serve error", zap.Error(err))
								health.ready.Store(false)
							}
						}()

						return nil
					},
					OnStop: func(ctx context.Context) error {
						logger.Info("stopping gRPC server...")
						grpcServer.GracefulStop()
						health.ready.Store(false)
						return nil
					},
				})
			},
		),
	)
}
