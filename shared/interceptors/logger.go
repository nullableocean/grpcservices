package interceptors

import (
	"context"
	"fmt"
	"time"

	"github.com/nullableocean/grpcservices/shared/xrequestid"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// логируем входящие запросы
func UnaryServerLogger(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		l := logger.With(
			zap.String("call_method", info.FullMethod),
			zap.String(xrequestid.XREQUEST_ID_KEY, xrequestid.GetFromIncomingCtx(ctx)),
		)

		l.Info("received grpc request")

		start := time.Now()
		resp, err = handler(ctx, req)

		l.Info("request handled", zap.Duration("duration", time.Since(start)), zap.Error(err))

		return resp, err
	}
}

// логируем исходящий запрос к grpc серверу
func UnaryClientLogger(logger *zap.Logger) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		start := time.Now()

		err := invoker(ctx, method, req, reply, cc, opts...)

		logger.Info("grpc client called",
			zap.String("called_method", method),
			zap.Duration("duration", time.Since(start)),
			zap.String(xrequestid.XREQUEST_ID_KEY, xrequestid.GetFromIncomingCtx(ctx)),
			zap.Error(err),
		)

		return err
	}
}

// ловим панику при отправке запроса
func UnaryClientPanicRecovery() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = status.Error(codes.Internal, fmt.Sprintf("sent grpc request panic: %v", r))
			}
		}()

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
