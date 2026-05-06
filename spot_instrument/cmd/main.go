package main

import (
	"log"

	"github.com/nullableocean/grpcservices/shared/logger"
	"github.com/nullableocean/grpcservices/spotinstrument/internal/app"
	"github.com/nullableocean/grpcservices/spotinstrument/internal/config"
)

func main() {
	cnf, err := config.NewConfig()
	if err != nil {
		log.Fatalln("failed init config", err)
	}

	zapLogger, err := logger.NewLogger(cnf.Log.Level, logger.Options{LogPath: cnf.Log.Path})
	if err != nil {
		log.Fatalln("failed init logger", err)
	}

	err = app.New(cnf, zapLogger).Run()
	if err != nil {
		log.Fatalln("failed start app", err)
	}
}
