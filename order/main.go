package main

import (
	"log"
	"main/api/orderpb"
	"main/api/spotpb"
	"main/order/client"
	"main/order/config"
	"main/order/logger"
	"main/order/server"
	"main/order/service"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	cnf, err := config.NewConfig()
	if err != nil {
		log.Fatalln("config init error:", err)
	}

	logger, err := logger.NewLogger(cnf)
	if err != nil {
		log.Fatalln("logger init error:", err)
	}

	spotGrpcConnect, err := grpc.NewClient(cnf.SpotServiceAddress+":"+cnf.Port,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		log.Fatalln("grpc connect to spot service error", err)
	}

	spotClient := client.NewSpotClient(spotpb.NewSpotInstrumentClient(spotGrpcConnect))

	userService := service.NewUserService()
	orderService := service.NewOrderService(spotClient, userService)
	orderServer := server.NewOrderServer(logger, orderService)

	lis, err := net.Listen("tcp", cnf.Address+":"+cnf.Port)
	if err != nil {
		log.Fatalln("start tcp listen error", err)
	}

	grpcServer := grpc.NewServer()
	orderpb.RegisterOrderServer(grpcServer, orderServer)

	logger.Info("ORDER SERVER STARTED....")
	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatalln("start serve grpc error", err)
	}
}
