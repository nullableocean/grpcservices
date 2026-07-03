package client

import (
	"context"

	orderv1 "github.com/nullableocean/grpcservices/api/gen/order/v1"
	"github.com/nullableocean/grpcservices/orderserviceclient/internal/dto"
	"google.golang.org/grpc/metadata"
)

func (c *Client) GetOrder(ctx context.Context, authToken string, dto *dto.GetOrderParams) (*Response, error) {
	if err := dto.Validate(); err != nil {
		return nil, err
	}

	req := &orderv1.GetOrderRequest{
		OrderUuid: dto.OrderUUID,
	}

	md := metadata.New(map[string]string{
		TokenMetadataKey: authToken,
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	response, err := c.connect.GetOrder(ctx, req)
	if err != nil {
		return nil, err
	}

	return &Response{
		Order: MapProtoOrderToOrder(response.Order),
	}, nil
}
