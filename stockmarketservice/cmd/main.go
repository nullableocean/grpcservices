package main

import (
	"log"

	"github.com/nullableocean/grpcservices/shared/logger"
	"github.com/nullableocean/grpcservices/stockmarketservice/internal/app"
	"github.com/nullableocean/grpcservices/stockmarketservice/internal/config"
	"go.uber.org/zap"
)

func main() {
	cnf, err := config.NewConfig()
	if err != nil {
		log.Fatalln("config init error: %w", err)
	}

	var zapLogger *zap.Logger
	if cnf.Log.LogToFile {
		zapLogger, err = logger.NewLoggerWithPath(cnf.Log.LogLevel, cnf.Log.LogPath)
	} else {
		zapLogger, err = logger.NewStdoutLogger(cnf.Log.LogLevel)
	}

	if err != nil {
		log.Fatalln("logger init error: %w", err)
	}

	err = app.Start(cnf, zapLogger)
	if err != nil {
		log.Fatalln(err)
	}
}
