package logger

import (
	"go.uber.org/zap"

	"github.com/nullableocean/grpcservices/order/config"
)

func NewLogger(cnf *config.Config) (*zap.Logger, error) {

	logConfig := zap.NewProductionConfig()
	logConfig.OutputPaths = []string{
		cnf.Log.LogPath,
		"stderr",
	}

	return logConfig.Build()
}
