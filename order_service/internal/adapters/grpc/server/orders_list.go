package server

import (
	"context"
	"time"

	modelsv1 "github.com/nullableocean/grpcservices/api/gen/models/v1"
	orderv1 "github.com/nullableocean/grpcservices/api/gen/order/v1"
	"github.com/nullableocean/grpcservices/orderservice/internal/adapters/grpc/mapping"
	"github.com/nullableocean/grpcservices/orderservice/internal/core/model"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (srv *OrderServer) OrdersList(ctx context.Context, req *orderv1.OrdersListRequest) (*orderv1.OrdersListResponse, error) {
	ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer("order_grpc_server").Start(ctx, "orders_list")
	defer span.End()

	user, err := srv.extractUserFromCtx(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "user not extracted from context")
	}

	srv.logger.Debug("received grpc request on orders list", zap.String("user_uuid", user.UUID))

	span.SetAttributes(attribute.String("user_uuid", user.UUID))

	filters, err := srv.extractFilters(req)
	if err != nil {
		span.AddEvent("failed extract filters")
		return nil, mapping.MapErrorToGrpcStatusError(err)
	}

	list, err := srv.orderService.OrdersList(ctx, user, filters)
	if err != nil {
		span.AddEvent("failed get orders list")
		return nil, mapping.MapErrorToGrpcStatusError(err)
	}

	resp := srv.createOrderListResponse(list)
	return resp, nil
}

func (srv *OrderServer) extractFilters(req *orderv1.OrdersListRequest) (model.OrderListFilter, error) {
	pbfilters := req.Filters

	var statuses []model.OrderStatus
	var markets []string
	var orderType *model.OrderType
	var createdFrom *time.Time
	var createdTo *time.Time

	if len(pbfilters.Statuses) > 0 {
		for _, s := range pbfilters.Statuses {
			statuses = append(statuses, mapping.MapProtoStatusToOrderStatus(s))
		}
	}

	if len(pbfilters.MarketUuids) > 0 {
		markets = pbfilters.MarketUuids
	}

	if pbfilters.Type != nil {
		t := mapping.MapProtoTypeToOrderType(*pbfilters.Type)
		orderType = &t
	}
	if pbfilters.CreatedFrom != nil {
		t := pbfilters.CreatedFrom.AsTime()
		createdFrom = &t
	}

	if pbfilters.CreatedTo != nil {
		t := pbfilters.CreatedTo.AsTime()
		createdTo = &t
	}

	cursor, err := model.DecodeTokenToCursor(req.PageToken)
	if err != nil {
		srv.logger.Error("failed decode pagination token", zap.Error(err))
	}

	return model.OrderListFilter{
		Statuses:    statuses,
		MarketUUIDs: markets,
		Type:        orderType,
		CreatedFrom: createdFrom,
		CreatedTo:   createdTo,
		PageCursor:  &cursor,
		Limit:       int(req.PageSize),
	}, nil
}

func (srv *OrderServer) createOrderListResponse(list model.OrderList) *orderv1.OrdersListResponse {
	pborders := make([]*modelsv1.Order, 0, len(list.List))

	for _, o := range list.List {
		pborders = append(pborders, mapping.MapOrderToProtoOrder(o))
	}

	var nextPageToken string

	if list.NextPageCursor != nil {
		nextPageToken = list.NextPageCursor.Encode()
	}

	return &orderv1.OrdersListResponse{
		Orders:        pborders,
		NextPageToken: nextPageToken,
	}
}
