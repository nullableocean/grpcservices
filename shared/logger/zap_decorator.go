package logger

import (
	"context"
	"errors"

	"github.com/nullableocean/grpcservices/shared/xrequestid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	TRACE_ID_KEY = "request_id"
)

type CtxZapLogger struct {
	*zap.Logger
}

func NewCtxZapLogger(logger *zap.Logger) (*CtxZapLogger, error) {
	if logger == nil {
		return nil, errors.New("argument *zap.Logger is nil")
	}

	return &CtxZapLogger{
		Logger: logger,
	}, nil
}

func (log *CtxZapLogger) Log(lvl zapcore.Level, msg string, fields ...zap.Field) {

	log.Logger.Log(lvl, msg, fields...)
}

// Debug logs a message at DebugLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (log *CtxZapLogger) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	log.Logger.Debug(msg, log.appendFieldsByContext(ctx, fields)...)
}

// Info logs a message at InfoLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (log *CtxZapLogger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	log.Logger.Info(msg, log.appendFieldsByContext(ctx, fields)...)
}

// Warn logs a message at WarnLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (log *CtxZapLogger) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	log.Logger.Warn(msg, log.appendFieldsByContext(ctx, fields)...)
}

// Error logs a message at ErrorLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (log *CtxZapLogger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	log.Logger.Error(msg, log.appendFieldsByContext(ctx, fields)...)
}

// DPanic logs a message at DPanicLevel. The message includes any fields
// passed at the log site, as well as any fields accumulated on the logger.
//
// If the logger is in development mode, it then panics (DPanic means
// "development panic"). This is useful for catching errors that are
// recoverable, but shouldn't ever happen.
func (log *CtxZapLogger) DPanic(ctx context.Context, msg string, fields ...zap.Field) {
	log.Logger.DPanic(msg, log.appendFieldsByContext(ctx, fields)...)
}

// Panic logs a message at PanicLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then panics, even if logging at PanicLevel is disabled.
func (log *CtxZapLogger) Panic(ctx context.Context, msg string, fields ...zap.Field) {
	log.Logger.Panic(msg, log.appendFieldsByContext(ctx, fields)...)
}

// Fatal logs a message at FatalLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then calls os.Exit(1), even if logging at FatalLevel is
// disabled.
func (log *CtxZapLogger) Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	log.Logger.Fatal(msg, log.appendFieldsByContext(ctx, fields)...)
}

// Named adds a new path segment to the logger's name. Segments are joined by
// periods. By default, Loggers are unnamed.
func (log *CtxZapLogger) Named(s string) *CtxZapLogger {
	return &CtxZapLogger{
		Logger: log.Logger.Named(s),
	}
}

// WithOptions clones the current Logger, applies the supplied Options, and
// returns the resulting Logger. It's safe to use concurrently.
func (log *CtxZapLogger) WithOptions(opts ...zap.Option) *CtxZapLogger {
	return &CtxZapLogger{
		Logger: log.Logger.WithOptions(opts...),
	}
}

// With creates a child logger and adds structured context to it. Fields added
// to the child don't affect the parent, and vice versa. Any fields that
// require evaluation (such as Objects) are evaluated upon invocation of With.
func (log *CtxZapLogger) With(fields ...zap.Field) *CtxZapLogger {
	return &CtxZapLogger{
		Logger: log.Logger.With(fields...),
	}
}

// WithLazy creates a child logger and adds structured context to it lazily.
//
// The fields are evaluated only if the logger is further chained with [With]
// or is written to with any of the log level methods.
// Until that occurs, the logger may retain references to objects inside the fields,
// and logging will reflect the state of an object at the time of logging,
// not the time of WithLazy().
//
// WithLazy provides a worthwhile performance optimization for contextual loggers
// when the likelihood of using the child logger is low,
// such as error paths and rarely taken branches.
//
// Similar to [With], fields added to the child don't affect the parent, and vice versa.
func (log *CtxZapLogger) WithLazy(fields ...zap.Field) *CtxZapLogger {
	return &CtxZapLogger{
		Logger: log.Logger.WithLazy(fields...),
	}
}

func (log *CtxZapLogger) appendFieldsByContext(ctx context.Context, fields []zap.Field) []zap.Field {
	if reqid := xrequestid.GetFromIncomingCtx(ctx); reqid != "" {
		fields = append(fields, zap.String(TRACE_ID_KEY, reqid))
	}

	return fields
}
