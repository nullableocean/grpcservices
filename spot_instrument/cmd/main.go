package main

import (
	"log"

	"github.com/nullableocean/grpcservices/shared/logger"
	"github.com/nullableocean/grpcservices/spotinstrument/internal/app"
	"github.com/nullableocean/grpcservices/spotinstrument/internal/config"
	"go.uber.org/zap"
)

func main() {
	cnf, err := config.NewConfig()
	if err != nil {
		log.Fatalln("failed init config", err)
	}

	var zapLogger *zap.Logger
	if cnf.Log.LogToFile {
		zapLogger, err = logger.NewLoggerWithPath(cnf.Log.LogLevel, cnf.Log.LogPath)
	} else {
		zapLogger, err = logger.NewStdoutLogger(cnf.Log.LogLevel)
	}

	if err != nil {
		log.Fatalln("failed init logger", err)
	}

	err = app.NewApp(cnf, zapLogger).Run()
	if err != nil {
		log.Fatalln("failed start app", err)
	}
}
