package orderserver

import (
	"context"
	"errors"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	orderv1 "github.com/nullableocean/grpcservices/api/gen/order/v1"
	typesv1 "github.com/nullableocean/grpcservices/api/gen/types/v1"
	"github.com/nullableocean/grpcservices/orderservice/internal/errs"
	"github.com/nullableocean/grpcservices/orderservice/internal/service/metrics"
	"github.com/nullableocean/grpcservices/orderservice/internal/service/order"
	"github.com/nullableocean/grpcservices/orderservice/internal/service/order/streamer"
	"github.com/nullableocean/grpcservices/orderservice/internal/transport/grpc/mapping"
)

type OrderServer struct {
	orderv1.UnimplementedOrderServer

	orderService   *order.OrderService
	orderProcessor order.Processor
	changeStreamer *streamer.ChangesStreamer

	metrics *metrics.OrderServiceMetrics
	logger  *zap.Logger
}

func NewOrderServer(orderService *order.OrderService, logger *zap.Logger, metrics *metrics.OrderServiceMetrics, changeStreamer *streamer.ChangesStreamer, orderProcessor order.Processor) *OrderServer {
	return &OrderServer{
		orderService:   orderService,
		changeStreamer: changeStreamer,
		orderProcessor: orderProcessor,

		metrics: metrics,
		logger:  logger,
	}
}

func (serv *OrderServer) CreateOrder(ctx context.Context, req *orderv1.CreateOrderRequest) (*orderv1.CreateOrderResponse, error) {
	orderCreatingData := mapping.MapCreateOrderRequestToOrderDto(req)

	serv.logger.Info("request create order",
		zap.String("user_id", orderCreatingData.UserUuid),
		zap.String("market_id", orderCreatingData.UserUuid),
		zap.Int64("quantity", orderCreatingData.Quantity),
		zap.String("price", orderCreatingData.Price.Decimal.String()),
	)

	// телеметрия, метрики
	ctx, span := otel.Tracer("order_service").Start(ctx, "create_order")
	defer span.End()

	start := time.Now()
	defer func() {
		serv.metrics.CalledCreateOrder(orderCreatingData.UserUuid, time.Since(start))
	}()

	createdOrder, err := serv.orderService.CreateOrder(ctx, orderCreatingData)
	if err != nil {
		span.AddEvent("create order error")

		serv.logger.Warn("failed create order", zap.Error(err))
		return nil, serv.getGrpcError(err)
	}

	span.AddEvent("order created")
	span.SetAttributes(attribute.String("order_uuid", createdOrder.UUID))

	serv.logger.Info("order created", zap.String("order_uuid", createdOrder.UUID))

	err = serv.orderProcessor.Process(ctx, createdOrder)
	if err != nil {
		span.AddEvent("failed order process")
		serv.logger.Warn("failed process order", zap.Error(err))
	}

	resp := mapping.MapDomainOrderToProtoResponse(createdOrder)
	return resp, nil
}

func (serv *OrderServer) GetOrderStatus(ctx context.Context, req *orderv1.GetStatusRequest) (*orderv1.GetStatusResponse, error) {
	serv.logger.Info("get order status request",
		zap.String("user_id", req.UserUuid),
		zap.String("order_id", req.OrderUuid),
	)

	defer func(uid, oid string) {
		serv.metrics.CalledGetStatus(uid, oid)
	}(req.UserUuid, req.OrderUuid)

	ctx, span := otel.Tracer("order_service").Start(ctx, "get_order_status")
	defer span.End()
	span.SetAttributes(attribute.String("order_uuid", req.OrderUuid))

	orderStatus, err := serv.orderService.GetOrderStatus(ctx, req.OrderUuid, req.UserUuid)
	if err != nil {
		span.AddEvent("get status error")
		serv.logger.Warn("error get order status from order service", zap.Error(err))

		if errors.Is(err, errs.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		if errors.Is(err, errs.ErrInvalidData) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &orderv1.GetStatusResponse{
		Status: typesv1.OrderStatus(orderStatus),
	}

	return resp, nil
}

func (serv *OrderServer) StreamOrderUpdates(req *orderv1.GetStatusRequest, stream grpc.ServerStreamingServer[orderv1.GetStatusResponse]) error {
	logger := serv.logger.With(
		zap.String("user_uuid", req.UserUuid),
		zap.String("order_uuid", req.OrderUuid),
	)
	logger.Info("request streaming on status updates ")

	orderUuid := req.OrderUuid
	userUuid := req.UserUuid

	ctx, span := otel.Tracer("order_service").Start(stream.Context(), "stream_order_status")
	defer span.End()
	span.SetAttributes(attribute.String("order_uuid", orderUuid))

	_, err := serv.orderService.FindOrder(ctx, orderUuid, userUuid)
	if err != nil {
		span.AddEvent("failed streaming request")
		logger.Warn("failed find order for request", zap.Error(err))

		return serv.getGrpcError(err)
	}

	sub, err := serv.changeStreamer.Sub(ctx, orderUuid)
	if err != nil {
		span.AddEvent("failed streaming request")
		logger.Warn("error subscription on update in order service", zap.Error(err))

		return serv.getGrpcError(err)
	}
	defer serv.changeStreamer.Dissub(ctx, orderUuid, sub.Id)

	statusesCh := streamer.WatchStatusChanges(ctx, sub)
	for statusChange := range statusesCh {
		logger.Info("get status changes", zap.String("new_status", statusChange.NewStatus.String()))

		err := stream.Send(&orderv1.GetStatusResponse{
			Status: typesv1.OrderStatus(statusChange.NewStatus),
		})
		if err != nil {
			serv.logger.Info("failed send to stream", zap.Error(err))
			break
		}
	}

	return nil
}

func (serv *OrderServer) getGrpcError(err error) error {
	if errors.Is(err, errs.ErrNotFound) {
		return status.Error(codes.NotFound, err.Error())
	}

	if errors.Is(err, errs.ErrAccessDenied) {
		return status.Error(codes.PermissionDenied, err.Error())
	}

	if errors.Is(err, errs.ErrInvalidData) {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	return status.Error(codes.Internal, err.Error())
}
