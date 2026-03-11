package app

import (
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/nullableocean/grpcservices/shared/intercepter"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func serverUnaryInterceptors(logger *zap.Logger, serverMetrics *grpc_prometheus.ServerMetrics) grpc.ServerOption {
	return grpc.ChainUnaryInterceptor(
		intercepter.UnaryServerPanicRecovery(logger),
		intercepter.UnaryServerLogger(logger),
		intercepter.UnaryServerTelemtry(),
		serverMetrics.UnaryServerInterceptor(),
	)
}

func serverStreamInterceptors(logger *zap.Logger, serverMetrics *grpc_prometheus.ServerMetrics) grpc.ServerOption {
	return grpc.ChainStreamInterceptor(
		intercepter.StreamServerPanicRecovery(logger),
		serverMetrics.StreamServerInterceptor(),
	)
}

func clientsInterceptors(logger *zap.Logger, clientMetrics *grpc_prometheus.ClientMetrics) grpc.DialOption {
	return grpc.WithChainUnaryInterceptor(
		intercepter.UnaryClientPanicRecovery(),
		intercepter.UnaryClientXReqId(),
		intercepter.UnaryClientXReqIdTelemtry(),
		clientMetrics.UnaryClientInterceptor(),
		intercepter.UnaryClientLogger(logger),
	)
}
