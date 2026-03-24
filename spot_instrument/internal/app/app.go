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
	spotv1 "github.com/nullableocean/grpcservices/api/gen/spot/v1"
	shared_auth "github.com/nullableocean/grpcservices/shared/auth"
	"github.com/nullableocean/grpcservices/shared/telemetry"
	"github.com/nullableocean/grpcservices/spotinstrument/internal/adapters/grpc/interceptors"
	"github.com/nullableocean/grpcservices/spotinstrument/internal/adapters/grpc/server"
	"github.com/nullableocean/grpcservices/spotinstrument/internal/adapters/metrics"
	"github.com/nullableocean/grpcservices/spotinstrument/internal/adapters/repository/postgres"
	"github.com/nullableocean/grpcservices/spotinstrument/internal/config"
	"github.com/nullableocean/grpcservices/spotinstrument/internal/core/services/spotinstrument"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type App struct {
	config *config.Config
	logger *zap.Logger

	grpc struct {
		server *grpc.Server
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

	// INIT SERVICES

	marketsRepo, err := postgres.NewMarketRepository(app.logger, app.postgres.pool)
	if err != nil {
		return fmt.Errorf("failed create market repository: %w", err)
	}

	metricsRecorder := metrics.NewSpotInstrumentRecorder(app.prometheus.reg)
	spotInstrument := spotinstrument.NewSpotInstrument(app.logger, marketsRepo, metricsRecorder)
	spotInstrumentServer := server.NewSpotInstrumentServer(app.logger, spotInstrument)

	spotv1.RegisterSpotInstrumentServer(app.grpc.server, spotInstrumentServer)
	// START

	errChan := make(chan error, 1)
	err = app.startGrpcServer(errChan)
	if err != nil {
		return err
	}
	app.startHttpServer(errChan)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case <-quit:
	case e := <-errChan:
		err = e
	}

	app.grpc.server.GracefulStop()
	app.http.server.Shutdown(context.Background())

	app.postgres.pool.Close()

	return err
}

func (app *App) setupGrpcServer() {
	app.grpc.server = grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		interceptors.ServerUnaryInterceptors(app.logger, app.prometheus.grpcMetricsSrv, shared_auth.NewHmacJwtAuth(app.config.Auth.JWT_SECRET)),
	)
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
