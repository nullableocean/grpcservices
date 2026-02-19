package main

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
	"google.golang.org/grpc/credentials/insecure"

	"github.com/nullableocean/grpcservices/api/orderpb"
	"github.com/nullableocean/grpcservices/api/spotpb"
	"github.com/nullableocean/grpcservices/order/client"
	"github.com/nullableocean/grpcservices/order/config"
	"github.com/nullableocean/grpcservices/order/logger"
	"github.com/nullableocean/grpcservices/order/seed"
	"github.com/nullableocean/grpcservices/order/server"
	"github.com/nullableocean/grpcservices/order/service/order"
	"github.com/nullableocean/grpcservices/order/service/user"
	"github.com/nullableocean/grpcservices/pkg/intercepter"
	"github.com/nullableocean/grpcservices/pkg/telemetry"
)

func start() error {
	cnf, err := config.NewConfig()
	if err != nil {
		return fmt.Errorf("config init error: %w", err)
	}

	logger, err := logger.NewLogger(cnf)
	if err != nil {
		return fmt.Errorf("logger init error: %w", err)
	}

	//telemetry
	ratioSampler := float64(1)
	shutdown, err := telemetry.InitTelemetryWithJaeger(cnf.App.Name, cnf.Telemetry.JaegerGrpcAddress, ratioSampler)
	if err != nil {
		return fmt.Errorf("init telemetry jaeger exporter error: %w", err)
	}
	defer shutdown(context.Background())

	//metrics
	serverMetrics := grpc_prometheus.NewServerMetrics()
	clientMetrics := grpc_prometheus.NewClientMetrics()

	promReg := prometheus.NewRegistry()
	promReg.MustRegister(serverMetrics, clientMetrics)

	//grpc server
	intersChain := grpc.ChainUnaryInterceptor(
		serverMetrics.UnaryServerInterceptor(),
		intercepter.UnaryServerLogger(logger),
		intercepter.UnaryServerPanicRecovery(),
	)
	grpcServer := grpc.NewServer(intersChain, grpc.StatsHandler(otelgrpc.NewServerHandler()))

	//grpc client
	clientInters := grpc.WithChainUnaryInterceptor(
		clientMetrics.UnaryClientInterceptor(),
		intercepter.UnaryClientXReqId(),
		intercepter.UnaryClientXReqIdTelemtry(),
		intercepter.UnaryClientPanicRecovery(),
	)
	spotGrpcConnect, err := grpc.NewClient(
		cnf.Spot.Address+":"+cnf.App.Port,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		clientInters,
	)

	if err != nil {
		return fmt.Errorf("grpc connect to spot service error: %w", err)
	}

	// services init
	spotClient := client.NewSpotClient(spotpb.NewSpotInstrumentClient(spotGrpcConnect))
	userService := user.NewUserService()
	orderService := order.NewOrderService(spotClient, userService)
	orderServer := server.NewOrderServer(logger, orderService)

	orderpb.RegisterOrderServer(grpcServer, orderServer)

	//listen init
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

	// тестовые юзеры если установлен флаг --seed
	if cnf.Seed.Need {
		seed.SeedUsers(logger, userService)
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
		logger.Info("order grpc service started", zap.String("address", cnf.App.Address+":"+cnf.App.Port))

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
