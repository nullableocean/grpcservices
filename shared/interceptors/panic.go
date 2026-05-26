package interceptors

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UnaryServerPanicRecovery ловит панику при обработке запроса
//
// stackDebugLines - количество строк из стека для дебага
func UnaryServerPanicRecovery(logger *zap.Logger, stackDebugLines int) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				msg := fmt.Sprintf("grpc request panic: %v", r)

				stack := debug.Stack()
				trimmedStack := strings.Join(strings.SplitN(string(stack), "\n", stackDebugLines), "\n")

				logger.Error("failed grpc request, got panic", zap.String("error", msg), zap.String("stack", trimmedStack))
				err = status.Error(codes.Internal, msg)
			}
		}()

		return handler(ctx, req)
	}
}

// StreamServerPanicRecovery ловит панику при обработке стрим-запроса
//
// stackDebugLines - количество строк из стека для дебага
func StreamServerPanicRecovery(logger *zap.Logger, stackDebugLines int) grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func() {
			if r := recover(); r != nil {
				msg := fmt.Sprintf("grpc stream panic: %v", r)

				stack := debug.Stack()
				trimmedStack := strings.Join(strings.SplitN(string(stack), "\n", stackDebugLines), "\n")

				logger.Error("failed grpc stream, got panic",
					zap.String("method", info.FullMethod),
					zap.String("error", msg),
					zap.String("stack", trimmedStack),
				)
				err = status.Error(codes.Internal, msg)
			}
		}()
		return handler(srv, stream)
	}
}
