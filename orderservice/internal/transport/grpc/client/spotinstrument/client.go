package spotinstrument

import (
	"context"

	spotv1 "github.com/nullableocean/grpcservices/api/gen/spot/v1"
	"github.com/nullableocean/grpcservices/shared/roles"

	"github.com/nullableocean/grpcservices/orderservice/internal/domain"
	"github.com/nullableocean/grpcservices/orderservice/internal/transport/mapping"
)

type SpotClient struct {
	client spotv1.SpotInstrumentClient
}

func NewSpotClient(client spotv1.SpotInstrumentClient) *SpotClient {
	return &SpotClient{
		client: client,
	}
}

func (cl *SpotClient) ViewMarkets(ctx context.Context, roles []roles.UserRole) ([]*domain.Market, error) {
	request := &spotv1.ViewMarketsRequest{
		UserRoles: mapping.MapUserRolesToProtoRoles(roles),
	}

	resp, err := cl.client.ViewMarkets(ctx, request)
	if err != nil {
		return nil, err
	}

	markets := mapping.MapProtoMarketsToDomainMarkets(resp.Markets)
	return markets, nil
}
