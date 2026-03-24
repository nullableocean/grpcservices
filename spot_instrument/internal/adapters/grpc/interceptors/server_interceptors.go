package interceptors

import (
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/nullableocean/grpcservices/shared/auth"
	shared_inters "github.com/nullableocean/grpcservices/shared/interceptors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func ServerUnaryInterceptors(logger *zap.Logger, serverMetrics *grpc_prometheus.ServerMetrics, authorizer auth.JwtAuthorizer) grpc.ServerOption {
	return grpc.ChainUnaryInterceptor(
		shared_inters.UnaryServerPanicRecovery(logger),
		shared_inters.UnaryServerLogger(logger),
		shared_inters.UnaryJwtAuthInterceptor(logger, authorizer),
		shared_inters.UnaryServerTelemtry(),
		serverMetrics.UnaryServerInterceptor(),
	)
}
