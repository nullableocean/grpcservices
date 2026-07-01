package client

import (
	"context"

	orderv1 "github.com/nullableocean/grpcservices/api/gen/order/v1"
	"github.com/nullableocean/grpcservices/orderserviceclient/internal/dto"
	"github.com/nullableocean/grpcservices/orderserviceclient/internal/model"
	"google.golang.org/grpc/metadata"
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
		Filters:   &orderv1.OrdersListFilter{},
		UserUuid:  dto.UserUUID,
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
