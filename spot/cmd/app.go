package main

import (
	"fmt"
	"main/api/spotpb"
	"main/pkg/intercepter"
	"main/spot/config"
	"main/spot/logger"
	"main/spot/server"
	"main/spot/service"
	"net"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
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

	// metrics
	grpcMetrics := grpc_prometheus.NewServerMetrics()

	promReg := prometheus.NewRegistry()
	promReg.MustRegister(grpcMetrics)

	//grpc init
	intersChain := grpc.ChainUnaryInterceptor(
		grpcMetrics.UnaryServerInterceptor(),
		intercepter.UnaryServerLoggerIntercepter(logger),
		intercepter.UnaryServerPanicRecoveryIntercepter(),
	)
	gprcServer := grpc.NewServer(intersChain)

	//register service
	spotInstrumentService := service.NewSpotInstrument()
	spotServer := server.NewSpotInstrumentServer(logger, spotInstrumentService)

	spotpb.RegisterSpotInstrumentServer(gprcServer, spotServer)

	// server start listen
	lis, err := net.Listen("tcp", cnf.Server.Address+":"+cnf.Server.Port)
	if err != nil {
		return fmt.Errorf("tcp listen error: %w", err)
	}

	logger.Info("SERVICE START LISTEN ON " + cnf.Server.Address + ":" + cnf.Server.Port)
	err = gprcServer.Serve(lis)
	if err != nil {
		return fmt.Errorf("grpc serve error: %w", err)
	}

	return nil
}
