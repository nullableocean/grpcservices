package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	spotv1 "github.com/nullableocean/grpcservices/api/gen/spot/v1"
	"github.com/nullableocean/grpcservices/shared/intercepter"
	"github.com/nullableocean/grpcservices/shared/telemetry"
	"github.com/nullableocean/grpcservices/spotinstrumentinstrument/internal/config"
	"github.com/nullableocean/grpcservices/spotinstrumentinstrument/internal/seed"
	"github.com/nullableocean/grpcservices/spotinstrumentinstrument/internal/server"
	guard "github.com/nullableocean/grpcservices/spotinstrumentinstrument/internal/service/auth"
	"github.com/nullableocean/grpcservices/spotinstrumentinstrument/internal/service/metrics"
	"github.com/nullableocean/grpcservices/spotinstrumentinstrument/internal/service/spot"
	"github.com/nullableocean/grpcservices/spotinstrumentinstrument/internal/store/ram"
)

func Start(cnf *config.Config, logger *zap.Logger) error {
	// telemetry
	ratioTracing := float64(1)
	shutdown, err := telemetry.InitTelemetryWithJaeger(cnf.App.Name, cnf.Telemetry.JaegerGrpcAddress, ratioTracing)
	if err != nil {
		return fmt.Errorf("telemtry init error: %w", err)
	}
	defer shutdown(context.Background())

	// metrics
	grpcMetrics := grpc_prometheus.NewServerMetrics()

	promReg := prometheus.NewRegistry()
	promReg.MustRegister(grpcMetrics)

	//grpc init
	intersChain := grpc.ChainUnaryInterceptor(
		intercepter.UnaryServerPanicRecovery(logger),
		intercepter.UnaryServerLogger(logger),
		intercepter.UnaryServerTelemtry(),
		grpcMetrics.UnaryServerInterceptor(),
	)
	grpcServer := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()), intersChain)

	//register service
	marketStore := ram.NewMarketStore()

	roleInspector := guard.NewRoleInspector()
	spotInstrumentService := spot.NewSpotInstrument(marketStore, roleInspector)

	spotMetrics := metrics.NewSpotMetrics(promReg)
	spotServer := server.NewSpotInstrumentServer(spotInstrumentService, logger, spotMetrics)

	spotv1.RegisterSpotInstrumentServer(grpcServer, spotServer)

	// server init listen

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(promReg, promhttp.HandlerOpts{}))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	httpServer := &http.Server{
		Addr:    ":" + cnf.Metrics.Port,
		Handler: mux,
	}

	if cnf.Seed {
		seed.SeedMarkets(logger, spotInstrumentService)
	}

	return upAndWaitShutdown(logger, cnf, grpcServer, httpServer)
}

// gracefull
func upAndWaitShutdown(logger *zap.Logger, cnf *config.Config, grpcServer *grpc.Server, httpServer *http.Server) error {
	var err error
	errChan := make(chan error, 1)

	go func() {
		logger.Info("start listen metrics http", zap.String("address", cnf.App.Address+":"+cnf.Metrics.Port))

		err := httpServer.ListenAndServe()
		if err != nil {
			errChan <- err
		}
	}()

	lis, err := net.Listen("tcp", cnf.App.Address+":"+cnf.App.Port)
	if err != nil {
		return fmt.Errorf("create listen tcp error: %w", err)
	}

	go func() {
		logger.Info("spot grpc service started", zap.String("address", cnf.App.Address+":"+cnf.App.Port))

		err = grpcServer.Serve(lis)
		if err != nil {
			errChan <- fmt.Errorf("start serve grpc error: %w", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case <-quit:
		grpcServer.GracefulStop()
		err = httpServer.Shutdown(context.Background())
	case e := <-errChan:
		err = e
	}

	return err
}
