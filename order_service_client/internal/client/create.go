package client

import (
	"context"

	"github.com/google/uuid"
	orderv1 "github.com/nullableocean/grpcservices/api/gen/order/v1"
	"github.com/nullableocean/grpcservices/orderserviceclient/internal/dto"
	"github.com/nullableocean/grpcservices/orderserviceclient/internal/model"
	"google.golang.org/grpc/metadata"
)

type Response struct {
	Order *model.Order
}

func (c *Client) CreateOrder(ctx context.Context, authToken string, dto *dto.CreateOrderParams) (*Response, error) {
	if err := dto.Validate(); err != nil {
		return nil, err
	}

	idemKey := dto.IdemKey
	if idemKey == "" {
		idemKey = uuid.NewString()
	}

	req := &orderv1.CreateOrderRequest{
		MarketUuid:     dto.MarketUUID,
		OrderType:      MapOrderTypeToProtoType(dto.Type),
		OrderSide:      MapOrderSideToProtoSide(dto.Side),
		Price:          MapDecimalToProtoMoney(dto.Price),
		Quantity:       MapDecimalToProtoDecimal(dto.Quantity),
		IdempotencyKey: idemKey,
	}

	md := metadata.New(map[string]string{
		TokenMetadataKey: authToken,
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	response, err := c.connect.CreateOrder(ctx, req)
	if err != nil {
		return nil, err
	}

	return &Response{
		Order: MapProtoOrderToOrder(response.Order),
	}, nil
}
