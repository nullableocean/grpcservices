package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/jackc/pgx/v5/pgxpool"
	orderv1 "github.com/nullableocean/grpcservices/api/gen/order/v1"
	spotv1 "github.com/nullableocean/grpcservices/api/gen/spot/v1"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/access"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/events/publishers"
	updatenotifier "github.com/nullableocean/grpcservices/orderservice/internal/adapters/events/publishers/update_notifier"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/grpc/client"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/grpc/interceptors"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/grpc/server"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/metrics"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/repository/postgres"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/repository/postgres/outbox"
	"github.com/nullableocean/grpcservices/orderservice/internal/config"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/services/order"
	shared_auth "github.com/nullableocean/grpcservices/shared/auth"
	"github.com/nullableocean/grpcservices/shared/telemetry"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type App struct {
	config *config.Config
	logger *zap.Logger

	grpc struct {
		server                *grpc.Server
		spotInstrumentConnect *grpc.ClientConn
	}

	postgres struct {
		pool *pgxpool.Pool
	}

	http struct {
		server *http.Server
	}

	prometheus struct {
		reg            *prometheus.Registry
		grpcMetricsSrv *grpc_prometheus.ServerMetrics
		grpcMetricsCl  *grpc_prometheus.ClientMetrics
	}

	kafka struct {
		updatesReader        *kafka.Reader
		marketsUpdatesReader *kafka.Reader
		createdEvWriter      *kafka.Writer
		dlqWriter            *kafka.Writer
	}
}

func NewApp(config *config.Config, logger *zap.Logger) *App {
	return &App{
		config: config,
		logger: logger,
	}
}

func (app *App) Run() error {
	//telemetry
	collectRatio := float64(1)
	shutdown, err := telemetry.InitTelemetryWithJaeger(app.config.App.Name, app.config.Telemetry.JaegerGrpcAddress, collectRatio)
	if err != nil {
		return fmt.Errorf("failed init telemetry jaeger exporter: %w", err)
	}
	defer shutdown(context.Background())

	err = app.setupPostgresPool()
	defer app.postgres.pool.Close()

	//metrics
	app.setupMetrics()

	//grpc server
	app.setupGrpcServer()

	//grpc clients connects
	err = app.setupGrpcConnects()
	if err != nil {
		return err
	}

	protoSpotClient := spotv1.NewSpotInstrumentClient(app.grpc.spotInstrumentConnect)

	spotInstrument := client.NewSpotInstrumentClient(app.logger, protoSpotClient)

	updatesNotifier := updatenotifier.NewUpdateNotifier(app.logger, updatenotifier.Options{})

	outboxWriter := outbox.NewOutboxWriter()
	orderRepository := postgres.NewOrderRepository(app.logger, app.postgres.pool, outboxWriter)

	accessService := access.NewRoleAccessService()
	metricsRecorder := metrics.NewPrometheusMetricsRecorder(app.prometheus.reg)
	orderService := order.NewOrderService(app.logger, orderRepository, spotInstrument, accessService, metricsRecorder)

	// events
	pubsBus := publishers.NewEventPublisherBus()
	pubsBus.Register(model.EVENT_ORDER_UPDATED, updatesNotifier)

	outboxRelay := outbox.NewRelay(app.logger, app.postgres.pool, pubsBus, outbox.Options{})

	orderServer := server.NewOrderServer(app.logger, orderService, updatesNotifier)
	orderv1.RegisterOrderServer(app.grpc.server, orderServer)

	errChan := make(chan error, 1)

	// servers
	err = app.startGrpcServer(errChan)
	if err != nil {
		return err
	}
	app.startHttpServer(errChan)

	// services
	outboxRelayCtx, cancelOutboxRelay := context.WithCancel(context.Background())
	outboxRelay.Start(outboxRelayCtx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case <-quit:
	case e := <-errChan:
		err = e
	}

	app.grpc.server.GracefulStop()
	app.http.server.Shutdown(context.Background())

	cancelOutboxRelay()
	app.postgres.pool.Close()

	return err
}

func (app *App) setupGrpcServer() {
	jwtAuthorizer := shared_auth.NewHmacJwtAuth(app.config.Auth.JWT_SECRET)

	app.grpc.server = grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		interceptors.ServerUnaryInterceptors(
			app.logger,
			app.prometheus.grpcMetricsSrv,
			jwtAuthorizer,
		),
		interceptors.ServerStreamInterceptors(app.logger, app.prometheus.grpcMetricsSrv, jwtAuthorizer),
	)
}

func (app *App) setupGrpcConnects() error {
	clientInterceptors := interceptors.ClientInterceptors(app.logger, app.prometheus.grpcMetricsCl)

	spotGrpcConnect, err := grpc.NewClient(
		app.config.Spot.Endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		clientInterceptors,
	)
	if err != nil {
		return fmt.Errorf("failed grpc connect to spot service: %w", err)
	}

	app.grpc.spotInstrumentConnect = spotGrpcConnect

	return nil
}

func (app *App) setupPostgresPool() error {
	pool, err := pgxpool.New(context.Background(), app.config.Postgres.DSN)
	if err != nil {
		return err
	}

	app.postgres.pool = pool
	return nil
}

func (app *App) setupMetrics() {
	app.prometheus.grpcMetricsSrv = grpc_prometheus.NewServerMetrics()
	app.prometheus.grpcMetricsCl = grpc_prometheus.NewClientMetrics()

	app.prometheus.reg = prometheus.NewRegistry()
	app.prometheus.reg.MustRegister(
		app.prometheus.grpcMetricsSrv,
		app.prometheus.grpcMetricsCl,
	)
}
