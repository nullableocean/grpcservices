package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	userv1 "github.com/nullableocean/grpcservices/api/gen/user/v1"
	"github.com/nullableocean/grpcservices/shared/intercepter"
	"github.com/nullableocean/grpcservices/shared/telemetry"
	"github.com/nullableocean/grpcservices/userservice/internal/auth"
	"github.com/nullableocean/grpcservices/userservice/internal/config"
	"github.com/nullableocean/grpcservices/userservice/internal/seed"
	"github.com/nullableocean/grpcservices/userservice/internal/service/user"
	"github.com/nullableocean/grpcservices/userservice/internal/store/ram"
	"github.com/nullableocean/grpcservices/userservice/internal/transport"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
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
	gprcServer := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()), intersChain)

	// service

	userStore := ram.NewUserStore()
	userService := user.NewUserService(logger, userStore, &auth.PasswordHasher{})
	userServer := transport.NewUserServer(logger, userService)

	userv1.RegisterUserServer(gprcServer, userServer)

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
		seed.SeedUsers(logger, userService)
	}

	return upAndWaitShutdown(logger, cnf, gprcServer, httpServer)
}

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
		logger.Info("user grpc service started", zap.String("address", cnf.App.Address+":"+cnf.App.Port))

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
