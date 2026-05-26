package logger

import (
	"io"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger
// levels "debug" "info" "warn" "error" "panic" "fatal"
func NewLogger(level string, outputs ...io.Writer) (*zap.Logger, error) {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, err
	}

	if len(outputs) == 0 {
		outputs = []io.Writer{os.Stdout}
	}

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.RFC3339NanoTimeEncoder

	wSyncs := make([]zapcore.WriteSyncer, 0, len(outputs))

	for _, outWr := range outputs {
		syncer := zapcore.AddSync(outWr)
		wSyncs = append(wSyncs, syncer)
	}

	syncer := zapcore.NewMultiWriteSyncer(wSyncs...)
	core := zapcore.NewCore(zapcore.NewJSONEncoder(encoderCfg), syncer, lvl)

	logger := zap.New(core, zap.AddCaller())

	return logger, nil
}
