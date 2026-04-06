package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/jackc/pgx/v5/pgxpool"
	spotv1 "github.com/nullableocean/grpcservices/api/gen/spot/v1"
	shared_auth "github.com/nullableocean/grpcservices/shared/auth"
	shared_inters "github.com/nullableocean/grpcservices/shared/interceptors"
	shared_telemetry "github.com/nullableocean/grpcservices/shared/telemetry"
	"github.com/nullableocean/grpcservices/spotinstrument/internal/adapters/grpc/server"
	"github.com/nullableocean/grpcservices/spotinstrument/internal/adapters/metrics"
	"github.com/nullableocean/grpcservices/spotinstrument/internal/adapters/repository/postgres"
	"github.com/nullableocean/grpcservices/spotinstrument/internal/config"
	"github.com/nullableocean/grpcservices/spotinstrument/internal/core/services/spotinstrument"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

type App struct {
	cfg    *config.Config
	logger *zap.Logger

	pgPool         *pgxpool.Pool
	metricsReg     *prometheus.Registry
	grpcMetricsSrv *grpc_prometheus.ServerMetrics
	grpcMetricsCl  *grpc_prometheus.ClientMetrics
	grpcServer     *grpc.Server
	httpServer     *http.Server

	closers []func() error
}

func New(cfg *config.Config, logger *zap.Logger) *App {
	return &App{
		cfg:    cfg,
		logger: logger,
	}
}

func (a *App) Run() error {
	if err := a.initTelemetry(); err != nil {
		return err
	}

	if err := a.initDB(); err != nil {
		return err
	}
	defer a.closeDB()

	if err := a.initMetrics(); err != nil {
		return err
	}

	if err := a.initGRPCServer(); err != nil {
		return err
	}

	if err := a.initServices(); err != nil {
		return err
	}

	errChan := make(chan error, 2)

	a.startHTTPServer(errChan)
	if err := a.startGRPCServer(errChan); err != nil {
		return err
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGQUIT)

	var listenError error
	select {
	case <-quit:
	case err := <-errChan:
		errs := make([]error, 0, len(errChan)+1)
		errs = append(errs, err)
		for range len(errChan) {
			errs = append(errs, <-errChan)
		}
		listenError = errors.Join(errs...)
	}

	a.logger.Info("shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), a.cfg.App.ShutdownTimeout)
	defer cancel()

	a.grpcServer.GracefulStop()

	if err := a.httpServer.Shutdown(ctx); err != nil {
		a.logger.Error("failed HTTP server shutdown error", zap.Error(err))
	}

	for _, closer := range a.closers {
		if err := closer(); err != nil {
			a.logger.Error("closer error", zap.Error(err))
		}
	}

	return listenError
}

func (a *App) initTelemetry() error {
	shutdown, err := shared_telemetry.InitOpenTelemtryGrpcProvider(
		a.cfg.App.Name,
		a.cfg.Telemetry.ExporterGrpcAddress,
		a.cfg.Telemetry.SampleRatio,
	)
	if err != nil {
		return fmt.Errorf("failed init telemetry: %w", err)
	}

	a.closers = append(a.closers, func() error { return shutdown(context.Background()) })

	return nil
}

func (a *App) initDB() error {
	pgCfg, err := pgxpool.ParseConfig(a.cfg.Postgres.DSN)
	if err != nil {
		return fmt.Errorf("parse postgres dsn: %w", err)
	}

	pgCfg.MaxConns = a.cfg.Postgres.MaxConns
	pgCfg.MinConns = a.cfg.Postgres.MinConns
	pgCfg.MaxConnLifetime = a.cfg.Postgres.MaxConnLifetime
	pgCfg.MaxConnIdleTime = a.cfg.Postgres.MaxConnIdleTime
	pgCfg.ConnConfig.ConnectTimeout = a.cfg.Postgres.ConnTimeout

	pool, err := pgxpool.NewWithConfig(context.Background(), pgCfg)
	if err != nil {
		return fmt.Errorf("failed create postgres pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return fmt.Errorf("failed ping postgres: %w", err)
	}

	a.pgPool = pool

	return nil
}

func (a *App) closeDB() {
	if a.pgPool != nil {
		a.pgPool.Close()
	}
}

func (a *App) initMetrics() error {
	a.grpcMetricsSrv = grpc_prometheus.NewServerMetrics()
	a.grpcMetricsCl = grpc_prometheus.NewClientMetrics()
	a.metricsReg = prometheus.NewRegistry()

	if err := a.metricsReg.Register(a.grpcMetricsSrv); err != nil {
		return fmt.Errorf("failed register server metrics: %w", err)
	}

	if err := a.metricsReg.Register(a.grpcMetricsCl); err != nil {
		return fmt.Errorf("failed register client metrics: %w", err)
	}

	return nil
}

func (a *App) initGRPCServer() error {
	jwtAuthorizer := shared_auth.NewHmacJwtAuth(a.cfg.Auth.JWTSecret)

	unaryInterceptors := grpc.ChainUnaryInterceptor(
		shared_inters.UnaryServerPanicRecovery(a.logger),
		shared_inters.UnaryServerLogger(a.logger),
		shared_inters.UnaryServerTelemtry(),
		a.grpcMetricsSrv.UnaryServerInterceptor(),
		shared_inters.ValidationUnaryInterceptor(),
		shared_inters.UnaryJwtAuthInterceptor(a.logger, jwtAuthorizer),
	)

	serverOpts := []grpc.ServerOption{
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    a.cfg.GRPC.Keepalive.Time,
			Timeout: a.cfg.GRPC.Keepalive.Timeout,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			PermitWithoutStream: a.cfg.GRPC.Keepalive.PermitWithoutStream,
		}),
		grpc.MaxRecvMsgSize(a.cfg.GRPC.ServerMaxRecvMsgSize),
		grpc.MaxSendMsgSize(a.cfg.GRPC.ServerMaxSendMsgSize),
		grpc.MaxConcurrentStreams(a.cfg.GRPC.ServerMaxConcurrentStreams),
		unaryInterceptors,
	}

	a.grpcServer = grpc.NewServer(serverOpts...)

	return nil
}

func (a *App) initServices() error {
	marketRepo, err := postgres.NewMarketRepository(a.logger, a.pgPool)
	if err != nil {
		return fmt.Errorf("failed create market repository: %w", err)
	}

	metricsRecorder := metrics.NewSpotInstrumentRecorder(a.metricsReg)
	spotInstrumentSvc := spotinstrument.NewSpotInstrument(a.logger, marketRepo, metricsRecorder)

	spotServer := server.NewSpotInstrumentServer(a.logger, spotInstrumentSvc)
	spotv1.RegisterSpotInstrumentServer(a.grpcServer, spotServer)

	return nil
}

func (a *App) startGRPCServer(errCh chan<- error) error {
	lis, err := net.Listen("tcp", net.JoinHostPort(a.cfg.App.Address, a.cfg.App.Port))
	if err != nil {
		return fmt.Errorf("create listener: %w", err)
	}

	go func() {
		a.logger.Info("gRPC server started", zap.String("addr", lis.Addr().String()))
		if err := a.grpcServer.Serve(lis); err != nil {
			a.logger.Error("failed gRPC server serve start", zap.Error(err))

			errCh <- err
		}
	}()

	return nil
}

func (a *App) startHTTPServer(errCh chan<- error) {
	mux := http.NewServeMux()
	mux.Handle(a.cfg.Metrics.Path, promhttp.HandlerFor(a.metricsReg, promhttp.HandlerOpts{}))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	a.httpServer = &http.Server{
		Addr:    net.JoinHostPort(a.cfg.App.Address, a.cfg.Metrics.Port),
		Handler: mux,
	}

	go func() {
		a.logger.Info("HTTP server started", zap.String("addr", a.httpServer.Addr))
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Error("failed HTTP server listen", zap.Error(err))

			errCh <- err
		}
	}()
}
