package logger

import (
	"errors"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewStdoutLogger
// levels "debug" "info" "warn" "error" "panic" "fatal"
func NewStdoutLogger(level string) (*zap.Logger, error) {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, err
	}
	config := zap.NewProductionConfig()
	config.Level = lvl
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	config.OutputPaths = []string{"stdout"}

	return config.Build()
}

// NewLoggerWithPath
// levels "debug" "info" "warn" "error" "panic" "fatal"
func NewLoggerWithPath(level string, logPath string) (*zap.Logger, error) {
	if logPath == "" {
		return nil, errors.New("empty path")
	}

	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, err
	}
	logConfig := zap.NewProductionConfig()
	logConfig.EncoderConfig.TimeKey = "timestamp"
	logConfig.EncoderConfig.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	logConfig.OutputPaths = []string{
		logPath,
		"stdout",
	}
	logConfig.Level = lvl

	logger, err := logConfig.Build()
	if err != nil {
		return nil, err
	}

	return logger, nil
}
