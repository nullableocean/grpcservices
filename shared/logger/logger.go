package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Options struct {
	LogPath string // path to log file, empty if dont need
}

// NewLogger
// levels "debug" "info" "warn" "error" "panic" "fatal"
func NewLogger(level string, opts Options) (*zap.Logger, error) {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, err
	}

	logConfig := zap.NewProductionConfig()
	logConfig.EncoderConfig.TimeKey = "timestamp"
	logConfig.EncoderConfig.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	logConfig.Level = lvl

	logConfig.OutputPaths = []string{"stdout"}
	if opts.LogPath != "" {
		logConfig.OutputPaths = append(logConfig.OutputPaths, opts.LogPath)
	}

	logger, err := logConfig.Build()
	if err != nil {
		return nil, err
	}

	return logger, nil
}
