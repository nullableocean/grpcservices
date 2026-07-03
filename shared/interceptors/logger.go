package interceptors

import (
	"context"
	"time"

	"github.com/nullableocean/grpcservices/shared/xrequestid"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	RESPONSE_STATUS   = "grpc_response_status"
	CALLED_METHOD_KEY = "called_method"
	CALL_DURATION_KEY = "duration"
	STACK_KEY         = "stack"
)

// логируем входящие запросы
func UnaryServerLogger(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		fields := []zap.Field{
			zap.String(CALLED_METHOD_KEY, info.FullMethod),
		}

		if reqid := xrequestid.GetFromIncomingCtx(ctx); reqid != "" {
			fields = append(fields, zap.String(xrequestid.XREQUEST_ID_KEY, reqid))
		}

		l := logger.With(
			fields...,
		)

		l.Debug("received grpc request")

		start := time.Now()
		resp, err = handler(ctx, req)

		l.Debug("request handled", zap.Duration(CALL_DURATION_KEY, time.Since(start)), zap.String(RESPONSE_STATUS, getGRPCStatusCode(err).String()), zap.Error(err))
		return resp, err
	}
}

// логируем исходящий запрос к grpc серверу
func UnaryClientLogger(logger *zap.Logger) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		start := time.Now()

		err := invoker(ctx, method, req, reply, cc, opts...)

		logger.Debug("grpc client called",
			zap.String(CALLED_METHOD_KEY, method),
			zap.Duration(CALL_DURATION_KEY, time.Since(start)),
			zap.String(xrequestid.XREQUEST_ID_KEY, xrequestid.GetFromIncomingCtx(ctx)),
			zap.String(RESPONSE_STATUS, getGRPCStatusCode(err).String()),
			zap.Error(err),
		)

		return err
	}
}

func getGRPCStatusCode(err error) codes.Code {
	if err == nil {
		return codes.OK
	}
	st, ok := status.FromError(err)
	if !ok {
		return codes.Unknown
	}
	return st.Code()
}
