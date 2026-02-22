package server

import (
	"context"
	"errors"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/nullableocean/grpcservices/api/orderpb"
	"github.com/nullableocean/grpcservices/order/service"
	"github.com/nullableocean/grpcservices/order/service/order"
)

type OrderServer struct {
	orderpb.UnimplementedOrderServer

	orderService *order.OrderService
	mapper       *OrderServerMapper

	logger *zap.Logger
}

func NewOrderServer(logger *zap.Logger, orderService *order.OrderService) *OrderServer {
	return &OrderServer{
		orderService: orderService,
		mapper:       &OrderServerMapper{},

		logger: logger,
	}
}

func (serv *OrderServer) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.CreateOrderResponse, error) {
	orderCreatingData := serv.mapper.MapCreateRequestToOrderDto(req)
	order, err := serv.orderService.CreateOrder(ctx, orderCreatingData)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		if errors.Is(err, service.ErrInvalidData) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := serv.mapper.MapOrderToPbResponse(order)
	return resp, nil
}

func (serv *OrderServer) GetOrderStatus(ctx context.Context, req *orderpb.GetStatusRequest) (*orderpb.GetStatusResponse, error) {
	orderStatus, err := serv.orderService.GetOrderStatus(ctx, req.OrderId, req.UserId)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		if errors.Is(err, service.ErrInvalidData) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &orderpb.GetStatusResponse{
		Status: orderpb.OrderStatus(orderStatus),
	}

	return resp, nil
}
