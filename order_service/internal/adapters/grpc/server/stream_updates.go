package server

import (
	orderv1 "github.com/nullableocean/grpcservices/api/gen/order/v1"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/grpc/mapping"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
	"github.com/nullableocean/grpcservices/shared/interceptors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (srv *OrderServer) StreamOrderUpdates(req *orderv1.GetUpdatesRequest, stream grpc.ServerStreamingServer[orderv1.UpdatesResponse]) error {
	ctx, span := otel.Tracer("order_grpc_server").Start(stream.Context(), "get_order_updates_stream")
	defer span.End()

	userUUID, exist := interceptors.UserUUIDFromContext(ctx)
	if !exist || userUUID == "" {
		return status.Error(codes.Unauthenticated, "user not found in context")
	}

	orderUUID := req.OrderUuid

	span.SetAttributes(attribute.String("user_uuid", userUUID))
	span.SetAttributes(attribute.String("order_uuid", orderUUID))

	logger := srv.logger.With(zap.String("user_uuid", userUUID), zap.String("order_uuid", orderUUID))
	logger.Debug("grpc received call for start update streaming")

	_, err := srv.orderService.GetOrder(ctx, orderUUID, userUUID)
	if err != nil {
		logger.Warn("failed find order for streaming updates")
		return srv.getGrpcError(err)
	}

	sub := srv.updateNotifier.Subscribe(ctx, orderUUID)
	defer sub.Close()

SEND_UPDATES:
	for update := range sub.Updates() {
		if update.Data.NewStatus != nil {
			err := stream.Send(srv.createUpdatesResponse(update.Data))
			if err != nil {
				span.AddEvent("failed send to stream")
				logger.Warn("failed send message to grpc stream", zap.Error(err))

				break SEND_UPDATES
			}
		}
	}

	logger.Debug("close stream")
	span.AddEvent("close stream")

	return nil
}

func (srv *OrderServer) createUpdatesResponse(data *model.EventUpdatedData) *orderv1.UpdatesResponse {
	if data.NewStatus != nil {
		return &orderv1.UpdatesResponse{
			Status:    mapping.MapOrderStatusToProtoStatus(*data.NewStatus),
			OldStatus: mapping.MapOrderStatusToProtoStatus(*data.OldStatus),
			UpdatedAt: timestamppb.New(data.UpdatedAt),
		}
	}

	return &orderv1.UpdatesResponse{}
}
