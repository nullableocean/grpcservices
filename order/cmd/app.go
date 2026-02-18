package main

import (
	"fmt"
	"main/api/orderpb"
	"main/api/spotpb"
	"main/order/client"
	"main/order/config"
	"main/order/logger"
	"main/order/server"
	"main/order/service"
	"main/pkg/intercepter"
	"net"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	//metrics

	serverMetrics := grpc_prometheus.NewServerMetrics()
	clientMetrics := grpc_prometheus.NewClientMetrics()

	promReg := prometheus.NewRegistry()
	promReg.MustRegister(serverMetrics, clientMetrics)

	//grpc server
	intersChain := grpc.ChainUnaryInterceptor(
		serverMetrics.UnaryServerInterceptor(),
		intercepter.UnaryServerLoggerIntercepter(logger),
		intercepter.UnaryServerPanicRecoveryIntercepter(),
	)
	grpcServer := grpc.NewServer(intersChain)

	//grpc client
	clientInters := grpc.WithChainUnaryInterceptor(
		clientMetrics.UnaryClientInterceptor(),
		intercepter.UnaryClientXRequestIdIntercepter(),
		intercepter.UnaryClientPanicRecoveryIntercepter(),
	)
	spotGrpcConnect, err := grpc.NewClient(
		cnf.Spot.Address+":"+cnf.Server.Port,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		clientInters,
	)

	if err != nil {
		return fmt.Errorf("grpc connect to spot service error: %w", err)
	}

	// services init
	spotClient := client.NewSpotClient(spotpb.NewSpotInstrumentClient(spotGrpcConnect))
	userService := service.NewUserService()
	orderService := service.NewOrderService(spotClient, userService)
	orderServer := server.NewOrderServer(logger, orderService)

	orderpb.RegisterOrderServer(grpcServer, orderServer)

	//listen init
	lis, err := net.Listen("tcp", cnf.Server.Address+":"+cnf.Server.Port)
	if err != nil {
		return fmt.Errorf("start tcp listen error: %w", err)
	}

	logger.Info("SERVICE START LISTEN ON " + cnf.Server.Address + ":" + cnf.Server.Port)
	err = grpcServer.Serve(lis)
	if err != nil {
		return fmt.Errorf("start serve grpc error: %w", err)
	}

	return nil
}
