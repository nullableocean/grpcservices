package server

import (
	orderv1 "github.com/nullableocean/grpcservices/api/gen/order/v1"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/grpc/mapping"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/ports"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/services/order"
	"go.uber.org/zap"
)

type OrderServer struct {
	orderv1.UnimplementedOrderServer

	orderService   order.Service
	updateNotifier ports.UpdateNotifier

	logger *zap.Logger
}

func NewOrderServer(l *zap.Logger, orderService order.Service, updateNotifier ports.UpdateNotifier) *OrderServer {
	return &OrderServer{
		orderService:   orderService,
		updateNotifier: updateNotifier,
		logger:         l,
	}
}

func (srv *OrderServer) getGrpcError(e error) error {
	return mapping.MapErrorToGrpcStatus(e)
}
