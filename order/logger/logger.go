package logger

import (
	"main/order/config"

	"go.uber.org/zap"
)

func NewLogger(cnf *config.Config) (*zap.Logger, error) {

	logConfig := zap.NewProductionConfig()
	logConfig.OutputPaths = []string{
		cnf.Log.LogPath,
		"stderr",
	}

	return logConfig.Build()
}
