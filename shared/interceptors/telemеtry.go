package interceptors

import (
	"context"

	"github.com/nullableocean/grpcservices/shared/xrequestid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// UnaryServerTelemetry извлекает из входящего контекста x-request-id или генерирует новый и добавляет его в атрибуты трейса
func UnaryServerTelemetry(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		span := trace.SpanFromContext(ctx)
		if span.IsRecording() {
			reqid := xrequestid.GetFromIncomingCtx(ctx)
			if reqid == "" {
				var xreqErr error
				reqid, xreqErr = xrequestid.NewXRequestId()
				if xreqErr != nil {
					logger.Error("failed create xrequestid for income request", zap.Error(err))
				}

				ctx = xrequestid.SetInOutCtx(reqid, ctx)
			}

			span.SetAttributes(attribute.String(xrequestid.XREQUEST_ID_KEY, reqid))
		}

		return handler(ctx, req)
	}
}

// UnaryClientXReqIdTelemetry добавляет x-request-id в атрибуты трейса
func UnaryClientXReqIdTelemetry() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		reqid := xrequestid.GetFromIncomingCtx(ctx)

		span := trace.SpanFromContext(ctx)
		if span.IsRecording() {
			span.SetAttributes(attribute.String(xrequestid.XREQUEST_ID_KEY, reqid))
		}

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// UnaryClientXReqId добавляет x-request-id в исходящий контекст
func UnaryClientXReqId(logger *zap.Logger) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		reqid := xrequestid.GetFromIncomingCtx(ctx)

		if reqid == "" {
			var err error
			ctx, err = xrequestid.CreateToOutCtx(ctx)
			if err != nil {
				logger.Error("failed create xrequestid for outgoing request", zap.Error(err))
			}
		} else {
			ctx = xrequestid.SetInOutCtx(reqid, ctx)
		}

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
