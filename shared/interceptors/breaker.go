package interceptors

import (
	"context"

	"github.com/sony/gobreaker"
	"google.golang.org/grpc"
)

func UnaryCircuitBreakerInterceptor(cb *gobreaker.CircuitBreaker, skipErr func(error) bool) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		var grpcErr error

		_, cbErr := cb.Execute(func() (interface{}, error) {
			grpcErr = invoker(ctx, method, req, reply, cc, opts...)

			if skipErr(grpcErr) {
				return nil, nil
			}

			return nil, grpcErr
		})

		if cbErr != nil {
			return cbErr
		}

		return grpcErr
	}
}
