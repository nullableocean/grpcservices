package interceptors

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ловим панику при обработке запроса
func UnaryServerPanicRecovery(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				msg := fmt.Sprintf("grpc request panic: %v", r)
				logger.Error("failed grpc request, got panic", zap.String("error", msg), zap.Stack("stack"))
				err = status.Error(codes.Internal, msg)
			}
		}()

		return handler(ctx, req)
	}
}

func StreamServerPanicRecovery(logger *zap.Logger) grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func() {
			if r := recover(); r != nil {
				msg := fmt.Sprintf("grpc stream panic: %v", r)
				logger.Error("failed grpc stream, got panic",
					zap.String("method", info.FullMethod),
					zap.String("error", msg),
					zap.Stack("stack"),
				)
				err = status.Error(codes.Internal, msg)
			}
		}()
		return handler(srv, stream)
	}
}
