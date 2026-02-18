package intercepter

import (
	"context"
	"fmt"

	"github.com/nullableocean/grpcservices/pkg/xrequestid"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Добавляет x-request-id в контекст запроса
func UnaryClientXRequestIdIntercepter() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx = xrequestid.SetNewRequestIdToCtx(ctx)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// логируем исходящий запрос к grpc серверу
func UnaryClientLoggerIntercepter(logger *zap.Logger) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		logger.Info("send grpc request",
			zap.String("method", method),
			zap.String(xrequestid.XREQUEST_ID_KEY, xrequestid.GetXRequestIdFromCtx(ctx)),
		)

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// ловим панику при отправке запроса
func UnaryClientPanicRecoveryIntercepter() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = status.Error(codes.Internal, fmt.Sprintf("sent grpc request panic: %v", r))
			}
		}()

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
