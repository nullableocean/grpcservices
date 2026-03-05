package main

import (
	"log"

	"github.com/nullableocean/grpcservices/orderservice/internal/app"
	"github.com/nullableocean/grpcservices/orderservice/internal/config"
	"github.com/nullableocean/grpcservices/shared/logger"
	"go.uber.org/zap"
)

func main() {
	cnf, err := config.NewConfig()
	if err != nil {
		log.Fatalf("config init error: %v\n", err)
	}

	var zapLogger *zap.Logger
	if cnf.Log.LogToFile {
		zapLogger, err = logger.NewLoggerWithPath(cnf.Log.LogLevel, cnf.Log.LogPath)
	} else {
		zapLogger, err = logger.NewStdoutLogger(cnf.Log.LogLevel)
	}

	if err != nil {
		log.Fatalf("logger init error: %v\n", err)
	}

	err = app.Run(cnf, zapLogger)
	if err != nil {
		log.Fatalf("start app error: %v\n", err)
	}
}
