package modules

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync/atomic"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	spotv1 "github.com/nullableocean/grpcservices/api/gen/spot/v1"
	shared_auth "github.com/nullableocean/grpcservices/shared/auth"
	"github.com/nullableocean/grpcservices/shared/health"
	shared_inters "github.com/nullableocean/grpcservices/shared/interceptors"
	"github.com/nullableocean/grpcservices/shared/logger"
	"github.com/nullableocean/grpcservices/spotinstrument/internal/adapters/grpc/server"
	"github.com/nullableocean/grpcservices/spotinstrument/internal/config"
	"github.com/nullableocean/grpcservices/spotinstrument/internal/core/services/spotinstrument"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

func GRPCServerModule() fx.Option {
	return fx.Options(
		fx.Provide(
			func(logger *zap.Logger, cfg *config.Config, serverMetrics *grpc_prometheus.ServerMetrics) *grpc.Server {
				jwtParser := shared_auth.NewHmacJwtParser(cfg.Auth.JWTSecret)

				unaryInterceptors := grpc.ChainUnaryInterceptor(
					shared_inters.UnaryServerPanicRecovery(logger, cfg.Log.StackLines),
					shared_inters.UnaryJwtAuthInterceptor(logger, jwtParser),
					shared_inters.UnaryServerLogger(logger),
					shared_inters.UnaryServerTelemetry(),
					serverMetrics.UnaryServerInterceptor(),
					shared_inters.ValidationUnaryInterceptor(logger),
				)

				streamInterceptors := grpc.ChainStreamInterceptor(
					shared_inters.StreamServerPanicRecovery(logger, cfg.Log.StackLines),
					serverMetrics.StreamServerInterceptor(),
					shared_inters.StreamJwtAuthInterceptor(logger, jwtParser),
				)

				return grpc.NewServer(
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
				)
			},
		),
	)
}

type GrpcServerHealth struct {
	ready atomic.Bool
}

func (h *GrpcServerHealth) HealthCheck() health.HealthCheck {
	return health.NewHealthCheck("grpc_server", func(ctx context.Context) error {
		if !h.ready.Load() {
			return errors.New("grpc server not ready")
		}
		return nil
	})
}

func GRPCRunnerModule() fx.Option {
	return fx.Options(
		fx.Provide(func() *GrpcServerHealth { return &GrpcServerHealth{} }),
		fx.Invoke(
			func(lc fx.Lifecycle, ctxLogger *logger.CtxZapLogger, logger *zap.Logger, cfg *config.Config, grpcServer *grpc.Server, spotService *spotinstrument.SpotInstrument, health *GrpcServerHealth) {
				spotServer := server.NewSpotInstrumentServer(ctxLogger, spotService)
				spotv1.RegisterSpotInstrumentServer(grpcServer, spotServer)

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
						health.ready.Store(false)
						grpcServer.GracefulStop()
						return nil
					},
				})
			},
		),
	)
}
