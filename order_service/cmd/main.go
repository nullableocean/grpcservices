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
		log.Fatalln("failed init config", err)
	}

	var zapLogger *zap.Logger
	if cnf.Log.LogToFile {
		zapLogger, err = logger.NewLoggerWithPath(cnf.Log.Level, cnf.Log.Path)
	} else {
		zapLogger, err = logger.NewStdoutLogger(cnf.Log.Level)
	}

	if err != nil {
		log.Fatalln("failed init logger", err)
	}

	err = app.New(cnf, zapLogger).Run()
	if err != nil {
		log.Fatalln("failed start app", err)
	}
}
