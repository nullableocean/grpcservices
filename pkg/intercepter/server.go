package intercepter

import (
	"context"
	"fmt"

	"github.com/nullableocean/grpcservices/pkg/xrequestid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func UnaryServerTelemtry() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		span := trace.SpanFromContext(ctx)
		if span.IsRecording() {
			reqid := xrequestid.GetFromIncomingCtx(ctx)

			span.SetAttributes(attribute.String(xrequestid.XREQUEST_ID_KEY, reqid))
		}

		return handler(ctx, req)
	}
}

// логируем входящие запросы
func UnaryServerLogger(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		logger.Info("got grpc request",
			zap.String("method", info.FullMethod),
			zap.String(xrequestid.XREQUEST_ID_KEY, xrequestid.GetFromIncomingCtx(ctx)),
		)

		return handler(ctx, req)
	}
}

// ловим панику при обработке запроса
func UnaryServerPanicRecovery() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = status.Error(codes.Internal, fmt.Sprintf("grpc request panic: %v", r))
			}
		}()

		return handler(ctx, req)
	}
}
