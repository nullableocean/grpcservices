package interceptors

import (
	"context"

	"google.golang.org/grpc"
)

type CircuitBreaker interface {
	Execute(context.Context, func() (interface{}, error)) (interface{}, error)
}

// UnaryCircuitBreakerInterceptor создаёт интерцептор для gRPC клиента, который оборачивает вызовы в circuit breaker
func UnaryCircuitBreakerInterceptor(cb CircuitBreaker) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		var grpcErr error

		_, cbErr := cb.Execute(ctx, func() (interface{}, error) {
			grpcErr = invoker(ctx, method, req, reply, cc, opts...)
			return nil, grpcErr
		})

		if cbErr != nil {
			return cbErr
		}

		return grpcErr
	}
}
