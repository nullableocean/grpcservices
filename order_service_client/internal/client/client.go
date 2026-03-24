package client

import (
	"context"
	"errors"
	"fmt"
	"io"

	orderv1 "github.com/nullableocean/grpcservices/api/gen/order/v1"
	"github.com/nullableocean/grpcservices/orderserviceclient/internal/dto"
	"google.golang.org/grpc/metadata"
)

const (
	TokenMetadataKey = "authorization"
)

type Client struct {
	connect orderv1.OrderClient
}

func NewClient(grpcAddr string) (*Client, error) {
	connect, err := newConnection(grpcAddr)
	if err != nil {
		return nil, err
	}

	return &Client{
		connect: connect,
	}, nil
}

func (c *Client) CreateOrder(ctx context.Context, token string, dto *dto.CreateOrderParameters) (*Response, error) {
	if err := dto.Validate(); err != nil {
		return nil, err
	}

	req := &orderv1.CreateOrderRequest{
		UserUuid:   dto.UserUUID,
		MarketUuid: dto.MarketUUID,
		OrderType:  MapOrderTypeToProtoType(dto.Type),
		OrderSide:  MapOrderSideToProtoSide(dto.Side),
		Price:      MapDecimalToProtoMoney(dto.Price),
		Quantity:   MapDecimalToProtoDeciaml(dto.Price),
	}

	md := metadata.New(map[string]string{
		TokenMetadataKey: token,
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	response, err := c.connect.CreateOrder(ctx, req)
	if err != nil {
		return nil, err
	}

	return &Response{
		NewOrderUuid: response.OrderUuid,
		Status:       MapProtoStatusToStatus(response.Status),
	}, nil
}

func (c *Client) StreamOrderUpdates(ctx context.Context, token string, dto *dto.StreamOrderUpdateDto) (<-chan *StreamData, error) {
	if err := dto.Validate(); err != nil {
		return nil, err
	}

	req := orderv1.GetStatusRequest{
		OrderUuid: dto.OrderUUID,
		UserUuid:  dto.UserUUID,
	}

	md := metadata.New(map[string]string{
		TokenMetadataKey: token,
	})

	ctx = metadata.NewOutgoingContext(ctx, md)

	stream, err := c.connect.StreamOrderUpdates(ctx, &req)
	if err != nil {
		return nil, fmt.Errorf("stream connection error: %w", err)
	}

	out := make(chan *StreamData, 1)

	go func() {
		defer close(out)

		for {
			resp, err := stream.Recv()
			data := &StreamData{}

			if err != nil {
				data.Err = err

				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					data.Err = fmt.Errorf("closed")
				}

				if errors.Is(err, io.EOF) {
					data.Err = fmt.Errorf("closed by server")
				}

				out <- data
				return
			}

			data.NewStatus = MapProtoStatusToStatus(resp.Status)
			out <- data
		}
	}()

	return out, nil
}
