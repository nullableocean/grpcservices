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

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/jackc/pgx/v5/pgxpool"
	orderv1 "github.com/nullableocean/grpcservices/api/gen/order/v1"
	spotv1 "github.com/nullableocean/grpcservices/api/gen/spot/v1"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/access"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/events/publishers"
	updatenotifier "github.com/nullableocean/grpcservices/orderservice/internal/adapters/events/publishers/update_notifier"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/grpc/client"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/grpc/server"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/metrics"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/repository/postgres"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/repository/postgres/outbox"
	"github.com/nullableocean/grpcservices/orderservice/internal/config"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/services/order"
	shared_auth "github.com/nullableocean/grpcservices/shared/auth"
	shared_inters "github.com/nullableocean/grpcservices/shared/interceptors"
	shared_telemetry "github.com/nullableocean/grpcservices/shared/telemetry"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sony/gobreaker"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

type App struct {
	cnf    *config.Config
	logger *zap.Logger

	pgPool         *pgxpool.Pool
	metricsReg     *prometheus.Registry
	grpcMetricsSrv *grpc_prometheus.ServerMetrics
	grpcMetricsCl  *grpc_prometheus.ClientMetrics
	grpcServer     *grpc.Server
	spotConn       *grpc.ClientConn
	httpServer     *http.Server

	orderService    *order.OrderService
	outboxRelay     *outbox.OutboxRelay
	updatesNotifier *updatenotifier.UpdateNotifier

	closers []func() error
}

func New(cfg *config.Config, logger *zap.Logger) *App {
	app := &App{
		cnf:    cfg,
		logger: logger,
	}

	return app
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

	if err := a.initGRPCClients(); err != nil {
		return err
	}
	defer a.closeGRPCClients()

	if err := a.initServices(); err != nil {
		return err
	}

	a.registerGRPCServer()

	errChan := make(chan error, 2)

	a.startHTTPServer(errChan)
	if err := a.startGRPCServer(errChan); err != nil {
		return err
	}

	outboxCtx, cancelOutbox := context.WithCancel(context.Background())
	defer cancelOutbox()
	go a.outboxRelay.Start(outboxCtx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGQUIT)

	var listensError error
	select {
	case <-quit:
	case err := <-errChan:
		errs := make([]error, 0, len(errChan)+1)
		errs = append(errs, err)

		for range len(errChan) {
			errs = append(errs, <-errChan)
		}

		listensError = errors.Join(errs...)
	}

	a.logger.Info("shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), a.cnf.App.ShutdownTimeout)
	defer cancel()

	a.grpcServer.GracefulStop()

	if err := a.httpServer.Shutdown(ctx); err != nil {
		a.logger.Error("HTTP server shutdown error", zap.Error(err))
	}

	for _, closer := range a.closers {
		if err := closer(); err != nil {
			a.logger.Error("close error", zap.Error(err))
		}
	}

	if listensError != nil {
		return listensError
	}

	return nil
}

func (a *App) initTelemetry() error {
	shutdown, err := shared_telemetry.InitOpenTelemtryGrpcProvider(
		a.cnf.App.Name,
		a.cnf.Telemetry.ExporterGrpcAddress,
		a.cnf.Telemetry.SampleRatio,
	)

	if err != nil {
		return fmt.Errorf("init telemetry: %w", err)
	}

	a.closers = append(a.closers, func() error {
		return shutdown(context.Background())
	})

	return nil
}

func (a *App) initDB() error {
	pgCnf, err := pgxpool.ParseConfig(a.cnf.Postgres.DSN)
	if err != nil {
		return fmt.Errorf("parse pg dsn: %w", err)
	}

	pgCnf.MaxConns = a.cnf.Postgres.MaxConns
	pgCnf.MinConns = a.cnf.Postgres.MinConns
	pgCnf.MaxConnLifetime = a.cnf.Postgres.MaxConnLifetime
	pgCnf.MaxConnIdleTime = a.cnf.Postgres.MaxConnIdleTime
	pgCnf.ConnConfig.ConnectTimeout = a.cnf.Postgres.ConnTimeout

	pool, err := pgxpool.NewWithConfig(context.Background(), pgCnf)
	if err != nil {
		return fmt.Errorf("create pg pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return fmt.Errorf("ping pg: %w", err)
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
		return fmt.Errorf("register server metrics: %w", err)
	}

	if err := a.metricsReg.Register(a.grpcMetricsCl); err != nil {
		return fmt.Errorf("register client metrics: %w", err)
	}

	return nil
}

func (a *App) initGRPCServer() error {
	jwtAuthorizer := shared_auth.NewHmacJwtAuth(a.cnf.Auth.JWTSecret)

	unaryInteseptors := grpc.ChainUnaryInterceptor(
		shared_inters.UnaryServerPanicRecovery(a.logger),               // panic recovery
		shared_inters.UnaryServerLogger(a.logger),                      // logging request
		shared_inters.UnaryServerTelemtry(),                            // telemetry tracing
		a.grpcMetricsSrv.UnaryServerInterceptor(),                      // request metrics
		shared_inters.UnaryJwtAuthInterceptor(a.logger, jwtAuthorizer), // authorize jwt
	)

	streamInterceptrors := grpc.ChainStreamInterceptor(
		shared_inters.StreamServerPanicRecovery(a.logger),               // panic recovery
		a.grpcMetricsSrv.StreamServerInterceptor(),                      // stream request metrics
		shared_inters.StreamJwtAuthInterceptor(a.logger, jwtAuthorizer), // authorize jwt
	)

	serverOpts := []grpc.ServerOption{
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    a.cnf.GRPC.Keepalive.Time,
			Timeout: a.cnf.GRPC.Keepalive.Timeout,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			PermitWithoutStream: a.cnf.GRPC.Keepalive.PermitWithoutStream,
		}),
		grpc.MaxRecvMsgSize(a.cnf.GRPC.ServerMaxRecvMsgSize),
		grpc.MaxSendMsgSize(a.cnf.GRPC.ServerMaxSendMsgSize),
		grpc.MaxConcurrentStreams(a.cnf.GRPC.ServerMaxConcurrentStreams),

		unaryInteseptors,
		streamInterceptrors,
	}

	a.grpcServer = grpc.NewServer(serverOpts...)

	return nil
}

func (a *App) initGRPCClients() error {
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "spotinstrument-breaker",
		MaxRequests: a.cnf.CircuitBreaker.MaxRequests,
		Interval:    a.cnf.CircuitBreaker.Interval,
		Timeout:     a.cnf.CircuitBreaker.Timeout,
	})

	retryOpts := []retry.CallOption{}

	interceptors := grpc.WithChainUnaryInterceptor(
		shared_inters.UnaryClientPanicRecovery(),         // panic
		shared_inters.UnaryClientXReqId(),                // set xrequestid
		shared_inters.UnaryClientXReqIdTelemtry(),        // save xreqid to telemetry
		a.grpcMetricsCl.UnaryClientInterceptor(),         // grpc client metrics
		shared_inters.UnaryClientLogger(a.logger),        // logging request
		shared_inters.UnaryClientJwtForwardInterceptor(), // forward jwt
		retry.UnaryClientInterceptor(retryOpts...),       // retry request
		shared_inters.UnaryCircuitBreakerInterceptor(cb), // breaker
	)

	conn, err := grpc.NewClient(a.cnf.Spot.Endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		interceptors,
	)

	if err != nil {
		return fmt.Errorf("failed create grpc client for spotinstrument: %w", err)
	}
	a.spotConn = conn

	return nil
}

func (a *App) closeGRPCClients() {
	if a.spotConn != nil {
		_ = a.spotConn.Close()
	}
}

func (a *App) initServices() error {
	spotProtoClient := spotv1.NewSpotInstrumentClient(a.spotConn)
	spotInstrument := client.NewSpotInstrumentClient(a.logger, spotProtoClient, client.Option{
		RequestTimeout: a.cnf.GRPC.ClientTimeout,
	})

	outboxWriter := outbox.NewOutboxWriter()
	orderRepo := postgres.NewOrderRepository(a.logger, a.pgPool, outboxWriter)

	accessService := access.NewRoleAccessService()
	metricsRecorder := metrics.NewPrometheusMetricsRecorder(a.metricsReg)
	a.orderService = order.NewOrderService(
		a.logger,
		orderRepo,
		spotInstrument,
		accessService,
		metricsRecorder,
	)

	pubBus := publishers.NewEventPublisherBus()
	a.updatesNotifier = updatenotifier.NewUpdateNotifier(a.logger, updatenotifier.Options{})
	pubBus.Register(model.EVENT_ORDER_UPDATED, a.updatesNotifier)

	a.outboxRelay = outbox.NewRelay(a.logger, a.pgPool, pubBus, outbox.Options{
		Interval:  a.cnf.Outbox.PollInterval,
		BatchSize: a.cnf.Outbox.BatchSize,
	})

	return nil
}

func (a *App) registerGRPCServer() {
	orderServer := server.NewOrderServer(a.logger, a.orderService, a.updatesNotifier)
	orderv1.RegisterOrderServer(a.grpcServer, orderServer)
}

func (a *App) startGRPCServer(errCh chan<- error) error {
	lis, err := net.Listen("tcp", net.JoinHostPort(a.cnf.App.Address, a.cnf.App.Port))
	if err != nil {
		return fmt.Errorf("create listener: %w", err)
	}

	go func() {
		a.logger.Info("gRPC server started", zap.String("addr", lis.Addr().String()))

		if err := a.grpcServer.Serve(lis); err != nil {
			a.logger.Error("failed gRPC server start", zap.Error(err))

			errCh <- err
		}
	}()

	return nil
}

func (a *App) startHTTPServer(errCh chan<- error) {
	mux := http.NewServeMux()
	mux.Handle(a.cnf.Metrics.Path, promhttp.HandlerFor(a.metricsReg, promhttp.HandlerOpts{}))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	a.httpServer = &http.Server{
		Addr:    net.JoinHostPort(a.cnf.App.Address, a.cnf.Metrics.Port),
		Handler: mux,
	}

	go func() {
		a.logger.Info("HTTP server started", zap.String("addr", a.httpServer.Addr))

		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Error("HTTP server error", zap.Error(err))

			errCh <- err
		}
	}()
}
