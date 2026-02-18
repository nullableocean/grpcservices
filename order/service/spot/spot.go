package spot

import (
	"context"

	"github.com/nullableocean/grpcservices/order/client"
	"github.com/nullableocean/grpcservices/order/domain"
	"github.com/nullableocean/grpcservices/pkg/roles"
)

type SpotInstrument struct {
	client *client.SpotClient
}

func NewSpotInstrument(spotClient *client.SpotClient) *SpotInstrument {
	return &SpotInstrument{
		client: spotClient,
	}
}

func (s *SpotInstrument) ViewMarkets(ctx context.Context, roles []roles.UserRole) ([]*domain.Market, error) {
	return s.client.ViewMarkets(ctx, roles)
}
