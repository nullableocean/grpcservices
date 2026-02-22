package server

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/nullableocean/grpcservices/api/orderpb"
	"github.com/nullableocean/grpcservices/order/service"
	"github.com/nullableocean/grpcservices/order/service/metrics"
	"github.com/nullableocean/grpcservices/order/service/order"
)

type OrderServer struct {
	orderpb.UnimplementedOrderServer

	orderService *order.OrderService
	mapper       *OrderServerMapper

	metrics *metrics.OrderServiceMetrics
	logger  *zap.Logger
}

func NewOrderServer(orderService *order.OrderService, logger *zap.Logger, metrics *metrics.OrderServiceMetrics) *OrderServer {
	return &OrderServer{
		orderService: orderService,
		mapper:       &OrderServerMapper{},

		metrics: metrics,
		logger:  logger,
	}
}

func (serv *OrderServer) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.CreateOrderResponse, error) {
	orderCreatingData := serv.mapper.MapCreateRequestToOrderDto(req)

	serv.logger.Info("create order request",
		zap.Int64("user_id", orderCreatingData.UserId),
		zap.Int64("market_id", orderCreatingData.MarketId),
		zap.Int64("quantity", orderCreatingData.Quantity),
		zap.Float64("price", orderCreatingData.Price),
	)
	start := time.Now()
	defer func() {
		serv.metrics.CalledCreateOrder(orderCreatingData.UserId, time.Since(start))
	}()

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
	serv.logger.Info("get order status request",
		zap.Int64("user_id", req.UserId),
		zap.Int64("order_id", req.OrderId),
	)
	defer func(uid, oid int64) {
		serv.metrics.CalledGetStatus(uid, oid)
	}(req.UserId, req.OrderId)

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
