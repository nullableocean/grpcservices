package interceptors

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ValidateRequest interface {
	Validate() error
}

func ValidationUnaryInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if v, ok := req.(ValidateRequest); ok {
			if err := v.Validate(); err != nil {
				logger.Error("grpc validation failed", zap.Error(err))
				return nil, status.Errorf(codes.InvalidArgument, "validation failed: %v", err)
			}
		}

		return handler(ctx, req)
	}
}
