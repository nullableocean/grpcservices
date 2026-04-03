package interceptors

import (
	"context"

	"github.com/nullableocean/grpcservices/shared/xrequestid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
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

// добавляет x-request-id в атрибуты трейса
func UnaryClientXReqIdTelemtry() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		reqid := xrequestid.GetFromIncomingCtx(ctx)

		span := trace.SpanFromContext(ctx)
		if span.IsRecording() {
			span.SetAttributes(attribute.String(xrequestid.XREQUEST_ID_KEY, reqid))
		}

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func UnaryClientXReqId() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		reqid := xrequestid.GetFromIncomingCtx(ctx)

		if reqid == "" {
			ctx = xrequestid.CreateToOutCtx(ctx)
		} else {
			ctx = xrequestid.SetInOutCtx(reqid, ctx)
		}

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
