package spot

import (
	"context"

	"github.com/nullableocean/grpcservices/orderservice/internal/domain"
	"github.com/nullableocean/grpcservices/orderservice/internal/transport/grpc/client/spotinstrument"
	"github.com/nullableocean/grpcservices/shared/roles"
)

type SpotInstrument struct {
	client *spotinstrument.SpotClient
}

func NewSpotInstrument(spotClient *spotinstrument.SpotClient) *SpotInstrument {
	return &SpotInstrument{
		client: spotClient,
	}
}

func (s *SpotInstrument) ViewMarkets(ctx context.Context, roles []roles.UserRole) ([]*domain.Market, error) {
	return s.client.ViewMarkets(ctx, roles)
}
