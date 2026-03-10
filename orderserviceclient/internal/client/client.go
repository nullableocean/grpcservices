package client

import (
	"context"
	"errors"
	"fmt"
	"io"

	orderv1 "github.com/nullableocean/grpcservices/api/gen/order/v1"
	typesv1 "github.com/nullableocean/grpcservices/api/gen/types/v1"
	"github.com/nullableocean/grpcservices/orderserviceclient/internal/dto"
	"github.com/nullableocean/grpcservices/shared/order"
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

func (c *Client) CreateOrder(ctx context.Context, dto *dto.CreateOrderDto) (*Response, error) {
	if err := dto.Validate(); err != nil {
		return nil, err
	}

	req := &orderv1.CreateOrderRequest{
		UserUuid:  dto.UserUuid,
		MarketId:  dto.MarketUuid,
		OrderType: typesv1.OrderType(dto.OrderType),
		Price:     mapDecimalToProtoMoney(dto.Price),
		Quantity:  dto.Price.IntPart(),
	}

	response, err := c.connect.CreateOrder(ctx, req)
	if err != nil {
		return nil, err
	}

	return &Response{
		NewOrderUuid: response.OrderUuid,
		Status:       order.OrderStatus(response.Status),
	}, nil
}

func (c *Client) StreamOrderUpdates(ctx context.Context, dto *dto.StreamOrderUpdateDto) (<-chan *StreamData, error) {
	if err := dto.Validate(); err != nil {
		return nil, err
	}

	req := orderv1.GetStatusRequest{
		OrderUuid: dto.OrderUuid,
		UserUuid:  dto.UserUuid,
	}
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

			data.NewStatus = order.OrderStatus(resp.Status)
			out <- data
		}
	}()

	return out, nil
}
