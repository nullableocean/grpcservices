package intercepter

import (
	"context"
	"fmt"
	"main/pkg/xrequestid"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// логируем входящие запросы
func UnaryServerLoggerIntercepter(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		logger.Info("got grpc request",
			zap.String("method", info.FullMethod),
			zap.String(xrequestid.XREQUEST_ID_KEY, xrequestid.GetXRequestIdFromCtx(ctx)),
		)

		return handler(ctx, req)
	}
}

// ловим панику при обработке запроса
func UnaryServerPanicRecoveryIntercepter() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		go func() {
			if r := recover(); r != nil {
				err = status.Error(codes.Internal, fmt.Sprintf("grpc request panic: %v", r))
			}
		}()

		return handler(ctx, req)
	}
}
