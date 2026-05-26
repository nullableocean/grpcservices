package main

import (
	"io"
	"log"
	"os"

	"github.com/nullableocean/grpcservices/shared/logger"
	"github.com/nullableocean/grpcservices/spotinstrument/internal/app"
	"github.com/nullableocean/grpcservices/spotinstrument/internal/config"
)

func main() {
	cnf, err := config.NewConfig()
	if err != nil {
		log.Fatalln("failed init config", err)
	}

	logOutputs := []io.Writer{os.Stdout}
	var logFile *os.File
	if cnf.Log.Path != "" {
		var err error
		logFile, err = os.OpenFile(cnf.Log.Path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("failed open log file: %v", err)
		}
		defer func() {
			err := logFile.Close()
			if err != nil {
				log.Fatalf("failed close log file: %v", err)
			}
		}()

		logOutputs = append(logOutputs, logFile)
	}

	zapLogger, err := logger.NewLogger(cnf.Log.Level, logOutputs...)
	if err != nil {
		log.Fatalf("failed init logger: %v", err)
	}
	defer zapLogger.Sync()

	err = app.New(cnf, zapLogger).Run()
	if err != nil {
		log.Fatalln("failed start app", err)
	}
}
