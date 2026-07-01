package client

import (
	"fmt"

	orderv1 "github.com/nullableocean/grpcservices/api/gen/order/v1"
	"github.com/nullableocean/grpcservices/shared/interceptors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

func newConnection(addr string) (orderv1.OrderClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()), connectionUnaryInterceptors())
	if err != nil {
		return nil, fmt.Errorf("grpc connection error: %w", err)

	}

	return orderv1.NewOrderClient(conn), nil
}

func connectionUnaryInterceptors() grpc.DialOption {
	return grpc.WithChainUnaryInterceptor(interceptors.UnaryClientXReqId())
}
