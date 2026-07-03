package modules

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/nullableocean/grpcservices/orderservice/internal/config"
	"github.com/nullableocean/grpcservices/shared/logger"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type outputLoggerFile struct {
	f *os.File
}

func LoggerModule() fx.Option {
	return fx.Options(
		fx.Provide(
			func(cfg *config.Config) (*zap.Logger, outputLoggerFile, error) {
				var outputs []io.Writer
				outputs = append(outputs, os.Stdout)
				outputFile := outputLoggerFile{}
				if cfg.Log.Path != "" {
					f, err := os.OpenFile(cfg.Log.Path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
					if err != nil {
						return nil, outputLoggerFile{}, fmt.Errorf("failed to open log file: %w", err)
					}
					outputs = append(outputs, f)
					outputFile.f = f
				}

				l, err := logger.NewZapLogger(cfg.Log.Level, outputs...)
				return l, outputFile, err
			},
		),
		fx.Invoke(
			func(lc fx.Lifecycle, logger *zap.Logger, outputFile outputLoggerFile) {
				lc.Append(fx.Hook{
					OnStop: func(ctx context.Context) error {
						_ = logger.Sync()
						if outputFile.f != nil {
							outputFile.f.Close()
						}

						return nil
					},
				})
			},
		),
	)
}
