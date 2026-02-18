package client

import (
	"context"

	"github.com/nullableocean/grpcservices/pkg/roles"

	"github.com/nullableocean/grpcservices/order/domain"

	"github.com/nullableocean/grpcservices/api/spotpb"
)

type SpotClient struct {
	client spotpb.SpotInstrumentClient
	mapper *SpotClientMapper
}

func NewSpotClient(client spotpb.SpotInstrumentClient) *SpotClient {
	return &SpotClient{
		client: client,
		mapper: &SpotClientMapper{},
	}
}

func (cl *SpotClient) ViewMarkets(ctx context.Context, roles []roles.UserRole) ([]*domain.Market, error) {
	request := &spotpb.ViewMarketsRequest{
		UserRoles: cl.mapper.ToPbRoles(roles),
	}

	resp, err := cl.client.ViewMarkets(ctx, request)
	if err != nil {
		return nil, err
	}

	markets := cl.mapper.FromPbToMarkets(resp.Markets)
	return markets, nil
}
