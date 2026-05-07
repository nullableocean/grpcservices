package server

import (
	orderv1 "github.com/nullableocean/grpcservices/api/gen/order/v1"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/grpc/mapping"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (srv *OrderServer) StreamOrderUpdates(req *orderv1.GetUpdatesRequest, stream grpc.ServerStreamingServer[orderv1.UpdatesResponse]) error {
	ctx, span := otel.Tracer("order_grpc_server").Start(stream.Context(), "get_order_status")
	defer span.End()

	userUUID := req.UserUuid
	orderUUID := req.OrderUuid

	span.SetAttributes(attribute.String("user_uuid", userUUID))
	span.SetAttributes(attribute.String("order_uuid", orderUUID))

	logger := srv.logger.With(zap.String("user_uuid", userUUID), zap.String("order_uuid", orderUUID))
	logger.Info("grpc received call for start update streaming")

	_, err := srv.orderService.GetOrder(ctx, orderUUID, userUUID)
	if err != nil {
		logger.Warn("failed find order for streaming updates")
		return srv.getGrpcError(err)
	}

	sub := srv.updateNotifier.Subscribe(ctx, orderUUID)
	defer sub.Close()

READ_UPDATES:
	for update := range sub.Updates() {
		if update.Data.NewStatus != nil {
			err := stream.Send(srv.mapUpdateDataToStreamMsg(update.Data))
			if err != nil {
				span.AddEvent("failed send to stream")
				logger.Warn("failed send message to grpc stream", zap.Error(err))

				break READ_UPDATES
			}
		}
	}

	logger.Info("close stream")
	span.AddEvent("close stream")

	return nil
}

func (srv *OrderServer) mapUpdateDataToStreamMsg(data *model.EventUpdatedData) *orderv1.UpdatesResponse {
	if data.NewStatus != nil {
		return &orderv1.UpdatesResponse{
			Status:    mapping.MapOrderStatusToProtoStatus(*data.NewStatus),
			OldStatus: mapping.MapOrderStatusToProtoStatus(*data.OldStatus),
			UpdatedAt: timestamppb.New(data.UpdatedAt),
		}
	}

	return &orderv1.UpdatesResponse{}
}
