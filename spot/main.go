package main

import (
	"log"
	"main/api/spotpb"
	"main/spot/config"
	"main/spot/logger"
	"main/spot/server"
	"main/spot/service"
	"net"

	"google.golang.org/grpc"
)

var ()

func main() {
	cnf, err := config.NewConfig()
	if err != nil {
		log.Fatalln("config init error:", err)
	}

	logger, err := logger.NewLogger(cnf)
	if err != nil {
		log.Fatalln("logger init error:", err)
	}

	spotInstrumentService := service.NewSpotInstrument()
	spotServer := server.NewSpotInstrumentServer(logger, spotInstrumentService)

	gprcServer := grpc.NewServer()
	spotpb.RegisterSpotInstrumentServer(gprcServer, spotServer)

	lis, err := net.Listen("tcp", cnf.Address+":"+cnf.Port)
	if err != nil {
		log.Fatalln("tcp listen error:", err)
	}

	logger.Info("SPOT SERVER STARTED....")
	err = gprcServer.Serve(lis)
	if err != nil {
		log.Fatalln("grpc serve error:", err)
	}
}
