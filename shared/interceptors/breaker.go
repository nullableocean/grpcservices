package interceptors

import (
	"context"

	"github.com/sony/gobreaker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func UnaryCircuitBreakerInterceptor(cb *gobreaker.CircuitBreaker) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		var grpcErr error

		_, cbErr := cb.Execute(func() (interface{}, error) {
			grpcErr = invoker(ctx, method, req, reply, cc, opts...)
			if grpcErr != nil && !isClientError(grpcErr) {
				return nil, grpcErr
			}

			return nil, nil
		})

		if cbErr != nil {
			return cbErr
		}

		return grpcErr
	}
}

func isClientError(err error) bool {
	if err == nil {
		return false
	}

	st, ok := status.FromError(err)
	if !ok {
		return false
	}

	code := st.Code()
	switch code {
	case codes.Canceled,
		codes.InvalidArgument,
		codes.NotFound,
		codes.AlreadyExists,
		codes.PermissionDenied,
		codes.FailedPrecondition,
		codes.OutOfRange,
		codes.Unauthenticated:
		return true
	default:
		return false
	}
}
