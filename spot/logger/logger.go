package logger

import (
	"main/spot/config"

	"go.uber.org/zap"
)

func NewLogger(cnf *config.Config) (*zap.Logger, error) {

	logConfig := zap.NewProductionConfig()
	logConfig.OutputPaths = []string{
		cnf.LogPath,
		"stderr",
	}

	return logConfig.Build()
}
