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

				logger.Error("failed grpc request, got panic", zap.String("error", msg), zap.String(STACK_KEY, trimmedStack))
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
					zap.String(CALLED_METHOD_KEY, info.FullMethod),
					zap.String(STACK_KEY, trimmedStack),
					zap.String("error", msg),
				)
				err = status.Error(codes.Internal, msg)
			}
		}()

		return handler(srv, stream)
	}
}

// UnaryClientPanicRecovery ловит панику при отправке запроса
func UnaryClientPanicRecovery(logger *zap.Logger, stackDebugLines int) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		defer func() {
			if r := recover(); r != nil {
				msg := fmt.Sprintf("grpc stream panic: %v", fmt.Sprintf("send grpc request panic: %v", r))
				stack := debug.Stack()
				trimmedStack := strings.Join(strings.SplitN(string(stack), "\n", stackDebugLines), "\n")

				logger.Error("failed grpc client request, got panic",
					zap.String(CALLED_METHOD_KEY, method),
					zap.String(STACK_KEY, trimmedStack),
					zap.String("error", msg),
				)

				err = status.Error(codes.Internal, msg)
			}
		}()

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
