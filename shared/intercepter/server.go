package intercepter

import (
	"context"
	"fmt"

	"github.com/nullableocean/grpcservices/shared/xrequestid"
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
func UnaryServerPanicRecovery(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				msg := fmt.Sprintf("grpc request panic: %v", r)
				logger.Warn("failed grpc request, got panic", zap.String("error", msg), zap.Stack("stack"))
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
				logger.Warn("failed grpc stream, got panic",
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
