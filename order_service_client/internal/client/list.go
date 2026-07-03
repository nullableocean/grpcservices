package client

import (
	"context"

	orderv1 "github.com/nullableocean/grpcservices/api/gen/order/v1"
	"github.com/nullableocean/grpcservices/orderserviceclient/internal/dto"
	"github.com/nullableocean/grpcservices/orderserviceclient/internal/model"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ListResponse struct {
	NextPageToken string
	Orders        []*model.Order
}

func (c *Client) ListOrders(ctx context.Context, authToken string, dto *dto.ListOrdersParams) (*ListResponse, error) {
	if err := dto.Validate(); err != nil {
		return nil, err
	}

	req := &orderv1.OrdersListRequest{
		Filters:   c.getFilters(dto.Filters),
		PageSize:  int32(dto.PageSize),
		PageToken: dto.NextPageToken,
	}

	md := metadata.New(map[string]string{
		TokenMetadataKey: authToken,
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	response, err := c.connect.OrdersList(ctx, req)
	if err != nil {
		return nil, err
	}

	orders := make([]*model.Order, 0, len(response.Orders))
	for _, pbo := range response.Orders {
		orders = append(orders, MapProtoOrderToOrder(pbo))
	}

	return &ListResponse{
		NextPageToken: response.NextPageToken,
		Orders:        orders,
	}, nil

}

func (c *Client) getFilters(filters dto.Filters) *orderv1.OrdersListFilter {
	pbfilters := &orderv1.OrdersListFilter{}

	pbfilters.MarketUuids = filters.MarketUuids

	if len(filters.Statuses) > 0 {
		for _, s := range filters.Statuses {
			pbfilters.Statuses = append(pbfilters.Statuses, MapStatusToProtoStatus(s))
		}
	}

	if filters.Type != nil {
		t := MapOrderTypeToProtoType(*filters.Type)
		pbfilters.Type = &t
	}

	if filters.CreatedFrom != nil {
		pbfilters.CreatedFrom = timestamppb.New(*filters.CreatedFrom)
	}

	if filters.CreatedTo != nil {
		pbfilters.CreatedTo = timestamppb.New(*filters.CreatedTo)
	}

	return pbfilters
}
