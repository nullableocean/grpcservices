package interceptors

import (
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	shared_inters "github.com/nullableocean/grpcservices/shared/interceptors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func ClientInterceptors(logger *zap.Logger, clientMetrics *grpc_prometheus.ClientMetrics) grpc.DialOption {
	return grpc.WithChainUnaryInterceptor(
		shared_inters.UnaryClientPanicRecovery(),
		shared_inters.UnaryClientXReqId(),
		shared_inters.UnaryClientXReqIdTelemtry(),
		clientMetrics.UnaryClientInterceptor(),
		shared_inters.UnaryClientLogger(logger),
		shared_inters.UnaryClientJwtForwardInterceptor(),
	)
}
